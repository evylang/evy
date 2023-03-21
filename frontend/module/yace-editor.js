//prettier-ignore
// https://github.com/petersolopov/yace - MIT licensed
// source: https://github.com/petersolopov/yace/blob/8ed1f99977c4db9bdd60db4e2f5ba4edfcfc1940/src/index.js
import {
  textareaStyles,
  preStyles,
  rootStyles,
  linesStyles,
} from "./styles.js";

class Yace {
  constructor(selector, options = {}) {
    if (!selector) {
      throw new Error("selector is not defined");
    }

    this.root =
      selector instanceof Node ? selector : document.querySelector(selector);

    if (!this.root) {
      throw new Error(`element with "${selector}" selector is not exist`);
    }

    const defaultOptions = {
      value: "",
      styles: {},
      plugins: [],
      highlighter: (value) => escape(value),
    };

    this.options = {
      ...defaultOptions,
      ...options,
    };

    this.init();
  }

  init() {
    this.textarea = document.createElement("textarea");
    this.pre = document.createElement("pre");

    Object.assign(this.root.style, rootStyles, this.options.styles);
    Object.assign(this.textarea.style, textareaStyles);
    Object.assign(this.pre.style, preStyles);

    this.root.appendChild(this.textarea);
    this.root.appendChild(this.pre);

    this.addTextareaEvents();
    this.update({ value: this.options.value });
    this.updateLines();
  }

  addTextareaEvents() {
    this.handleInput = (event) => {
      const textareaProps = runPlugins(this.options.plugins, event);
      this.update(textareaProps);
    };

    this.handleKeydown = (event) => {
      const textareaProps = runPlugins(this.options.plugins, event);
      this.update(textareaProps);
    };

    this.textarea.addEventListener("input", this.handleInput);
    this.textarea.addEventListener("keydown", this.handleKeydown);
  }

  update(textareaProps) {
    const { value, selectionStart, selectionEnd } = textareaProps;
    // should be before updating selection otherwise selection will be lost
    if (value != null) {
      this.textarea.value = value;
    }

    this.textarea.selectionStart = selectionStart;
    this.textarea.selectionEnd = selectionEnd;

    if (value === this.value || value == null) {
      return;
    }

    this.value = value;

    const highlighted = this.options.highlighter(value);
    this.pre.innerHTML = highlighted + "<br/>";

    this.updateLines();

    if (this.updateCallback) {
      this.updateCallback(value);
    }
  }

  updateLines() {
    if (!this.options.lineNumbers) {
      return;
    }

    if (!this.lines) {
      this.lines = document.createElement("pre");
      this.root.appendChild(this.lines);
      Object.assign(this.lines.style, linesStyles);
    }

    const lines = this.value.split("\n");
    const length = lines.length.toString().length;

    this.root.style.paddingLeft = `${length + 1}ch`;

    this.lines.innerHTML = lines
      .map((line, number) => {
        // prettier-ignore
        const lineNumber = `<span class="yace-line" style="position: absolute; opacity: .3; left: 0">${1 + number}</span>`
        // prettier-ignore
        const lineText = `<span style="color: transparent; pointer-events: none">${escape(line)}</span>`;
        return `${lineNumber}${lineText}`;
      })
      .join("\n");
  }

  destroy() {
    this.textarea.removeEventListener("input", this.handleInput);
    this.textarea.removeEventListener("keydown", this.handleKeydown);
  }

  onUpdate(callback) {
    this.updateCallback = callback;
  }
}

function runPlugins(plugins, event) {
  const { value, selectionStart, selectionEnd } = event.target;

  return plugins.reduce(
    (acc, plugin) => {
      return {
        ...acc,
        ...plugin(acc, event),
      };
    },
    { value, selectionStart, selectionEnd }
  );
}

function escape(unsafe) {
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

export default Yace;

// source: https://github.com/petersolopov/yace/blob/8ed1f99977c4db9bdd60db4e2f5ba4edfcfc1940/src/plugins/isKey.js
const CODES = {
  backspace: 8,
  tab: 9,
  enter: 13,
  shift: 16,
  control: 17,
  alt: 18,
  pause: 19,
  capslock: 20,
  escape: 27,
  " ": 32,
  pageup: 33,
  pagedown: 34,
  end: 35,
  home: 36,
  arrowleft: 37,
  arrowup: 38,
  arrowright: 39,
  arrowdown: 40,
  insert: 45,
  delete: 46,
  meta: 91,
  numlock: 144,
  scrolllock: 145,
  ";": 186,
  "=": 187,
  ",": 188,
  "-": 189,
  ".": 190,
  "/": 191,
  "`": 192,
  "[": 219,
  "\\": 220,
  "]": 221,
  "'": 222,

  // aliases
  add: 187,
};

const IS_MAC =
  typeof window != "undefined" &&
  /Mac|iPod|iPhone|iPad/.test(window.navigator.platform);

const MODIFIERS = {
  alt: "altKey",
  control: "ctrlKey",
  meta: "metaKey",
  shift: "shiftKey",
  "ctrl/cmd": IS_MAC ? "metaKey" : "ctrlKey",
};

function toKeyCode(name) {
  return CODES[name] || name.toUpperCase().charCodeAt(0);
}

function isKey(string, event) {
  const keys = string.split("+").reduce(
    (acc, key) => {
      if (MODIFIERS[key]) {
        acc.modifiers[MODIFIERS[key]] = true;
        return acc;
      }

      return {
        ...acc,
        keyCode: toKeyCode(key),
      };
    },
    {
      modifiers: {
        altKey: false,
        ctrlKey: false,
        metaKey: false,
        shiftKey: false,
      },
      keyCode: null,
    }
  );

  const hasModifiers = Object.keys(keys.modifiers).every((key) => {
    const value = keys.modifiers[key];
    return value ? event[key] : !event[key];
  });

  const hasKey = keys.keyCode ? event.which === keys.keyCode : true;

  return hasModifiers && hasKey;
}

export default isKey;

// https://github.com/petersolopov/yace/blob/8ed1f99977c4db9bdd60db4e2f5ba4edfcfc1940/src/plugins/history.js
import isKey from "./isKey.js";

function history() {
  let stack = [];
  let activeIndex = null;

  const initHistory = (record) => {
    activeIndex = 0;
    stack.push(record);
  };

  const rewriteHistory = (record) => {
    stack = stack.slice(0, activeIndex + 1);
    stack.push(record);
    activeIndex = stack.length - 1;
  };

  const shouldRecord = (record) => {
    return (
      stack[activeIndex].value !== record.value ||
      stack[activeIndex].selectionStart !== record.selectionStart ||
      stack[activeIndex].selectionEnd !== record.selectionEnd
    );
  };

  return (textareaProps, event) => {
    if (event.type === "keydown") {
      if (isKey("ctrl/cmd+z", event)) {
        event.preventDefault();

        if (activeIndex !== null) {
          // after applying all plugins it can be new props
          if (shouldRecord(textareaProps)) {
            stack.push(textareaProps);
            activeIndex++;
          }

          const newActiveIndex = Math.max(0, activeIndex - 1);
          activeIndex = newActiveIndex;
          return stack[newActiveIndex];
        }
      }

      if (isKey("ctrl/cmd+shift+z", event)) {
        event.preventDefault();

        if (activeIndex !== null) {
          const newActiveIndex = Math.min(stack.length - 1, activeIndex + 1);
          activeIndex = newActiveIndex;
          return stack[newActiveIndex];
        }
      }

      if (activeIndex === null) {
        initHistory(textareaProps);
        return;
      }

      if (shouldRecord(textareaProps)) {
        rewriteHistory(textareaProps);
        return;
      }
    }

    if (event.type === "input") {
      if (activeIndex === null) {
        initHistory(textareaProps);
      } else {
        rewriteHistory(textareaProps);
      }
    }
  };
}

export default history;

// source: https://github.com/petersolopov/yace/blob/8ed1f99977c4db9bdd60db4e2f5ba4edfcfc1940/src/plugins/preserveIndent.js

import isKey from "./isKey.js";

const preserveIndent = () => (textareaProps, event) => {
  const { value, selectionStart, selectionEnd } = textareaProps;

  if (!isKey("enter", event)) {
    return;
  }

  if (event.type !== "keydown") {
    return;
  }

  const currentLineNumber =
    value.substring(0, selectionStart).split("\n").length - 1;

  const lines = value.split("\n");
  const currentLine = lines[currentLineNumber];
  const matches = /^\s+/.exec(currentLine);

  if (!matches) {
    return;
  }

  event.preventDefault();
  const indent = matches[0];
  const newLine = "\n";

  const inserted = newLine + indent;

  const newValue =
    value.substring(0, selectionStart) +
    inserted +
    value.substring(selectionEnd);

  return {
    value: newValue,
    selectionStart: selectionStart + inserted.length,
    selectionEnd: selectionStart + inserted.length,
  };
};

export default preserveIndent;

// source: https://github.com/petersolopov/yace/blob/8ed1f99977c4db9bdd60db4e2f5ba4edfcfc1940/src/plugins/tab.js

import isKey from "./isKey.js";

const tab = (tabCharacter = "  ") => (textareaProps, event) => {
  const { value, selectionStart, selectionEnd } = textareaProps;

  if (event.type !== "keydown") {
    return;
  }

  if (isKey("shift+tab", event)) {
    event.preventDefault();
    const linesBeforeCaret = value.substring(0, selectionStart).split("\n");
    const startLine = linesBeforeCaret.length - 1;
    const endLine = value.substring(0, selectionEnd).split("\n").length - 1;
    const nextValue = value
      .split("\n")
      .map((line, i) => {
        if (i >= startLine && i <= endLine && line.startsWith(tabCharacter)) {
          return line.substring(tabCharacter.length);
        }

        return line;
      })
      .join("\n");

    if (value !== nextValue) {
      const startLineText = linesBeforeCaret[startLine];

      return {
        value: nextValue,
        // Move the start cursor if first line in selection was modified
        // It was modified only if it started with a tab
        selectionStart: startLineText.startsWith(tabCharacter)
          ? selectionStart - tabCharacter.length
          : selectionStart,
        // Move the end cursor by total number of characters removed
        selectionEnd: selectionEnd - (value.length - nextValue.length),
      };
    }

    return;
  }

  if (isKey("tab", event)) {
    event.preventDefault();
    if (selectionStart === selectionEnd) {
      const updatedSelection = selectionStart + tabCharacter.length;
      const newValue =
        value.substring(0, selectionStart) +
        tabCharacter +
        value.substring(selectionEnd);

      return {
        value: newValue,
        selectionStart: updatedSelection,
        selectionEnd: updatedSelection,
      };
    }

    const linesBeforeCaret = value.substring(0, selectionStart).split("\n");
    const startLine = linesBeforeCaret.length - 1;
    const endLine = value.substring(0, selectionEnd).split("\n").length - 1;

    return {
      value: value
        .split("\n")
        .map((line, i) => {
          if (i >= startLine && i <= endLine) {
            return tabCharacter + line;
          }

          return line;
        })
        .join("\n"),
      selectionStart: selectionStart + tabCharacter.length,
      selectionEnd:
        selectionEnd + tabCharacter.length * (endLine - startLine + 1),
    };
  }
};

export default tab;
