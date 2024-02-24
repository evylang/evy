# Evy Contributing Guide

Thank you for wanting to help out with Evy 🙏. There are many ways to contribute:

- Share your creation with us! We love to see what you've made with Evy. Post
  it on Evy's [Discord], send us an [email] or share it anonymously with the
  [web form].
- Join our Discord community. Ask questions, help others, or just chat about
  Evy stuff. This is the best place to get started!
- Report issues. Did you find a bug? Have a neat idea for Evy? Create an
  [bug report] on GitHub or message us on [Discord].
- Contribute to the codebase. If you're experienced with Go or frontend
  development and want to help build Evy itself, please follow the details
  below.

Read our [Code of Conduct](CODE_OF_CONDUCT.md) to keep our community safe and welcoming.

[Discord]: https://discord.com/channels/1008553546058313738/1008553546582605857
[email]: mailto:evy@evy.dev
[web form]: https://forms.gle/n6KLGDBmAjTc7Z6NA

## Contribute to the Codebase

Our goal is to keep Evy's codebase as clear and understandable as the language
itself. Evy is designed to be a learning experience in every way – including
how the code itself is written. Our main goal is to make programming
accessible and inviting, and we believe the project's code should reflect
that. This means we need to be disciplined and thoughtful when making changes
to the Evy repository, starting with how we put together commit messages and
pull requests.

### Commit Messages

Structure:

- **Title**: component: Verb phrase completing "This change modifies Evy to..."
- **Body**: Explain the WHY and WHAT, wrap lines at 72 characters.

Example:

```
vm: Create vm package

Create a vm package that executes bytecode created by the compiler. Currently
it can only support the simplest evy program: an inferred type declaration
from a constant, followed by an assignment. We are putting structure in place
here which we will expand.
```

Resource: Good commit messages: https://cbea.ms/git-commit/

### Pull Requests

Your PR description becomes the merge commit message when your PR is merged.
Take the same care with it as you do with your commit messages!

#### Key Idea: Make Your Code Reviewer's Life Easier

- **Tell a Story:** Use multiple commits to break your changes into logical steps. Each step should work on its own (make ci will check this!).
- **Tidy History:** Before submitting, clean up your commits. This makes it easier for reviewers to follow your thought process.
- **Catch Your Own Mistakes:** Do a self-review before asking others. It'll save everyone time!

**Why This Matters:** You're presenting your work. Clear, organized work is
easier to understand and helps reviewers give you the best feedback
possible.

We use a [merge script] that formats the merge commits and adds an emoji -
this is for the merge commit only. To see examples, run the following
commands on the evy repo:

- `git log --oneline`
- `git log --oneline --graph` shows the structure more clearly and
- `git log --oneline --first-parent` shows merge-commits / PR titles only

If you have non-permanent comments in the PR description, e.g.: "This PR is
temporarily deployed at https://evystage.dev", post them past a `---`, those
sections do not get copied into the merge commit.

Use `Fixes` or `Updates` keywords with issue numbers after `---` in your PR
description to automatically close or reference GitHub issues.

[merge script]: https://github.com/foxygoat/git-scripts/blob/master/git-pr-merge

### Go code

Try to follow the best practices outlined [Go Code Review Comments] and
[Effective Go].

#### godoc comments in code

Wrap at 72 char.

#### Filenames

Don't use `-` or `_` in _.go file names unless used with intention, e.g.
`_\_test.go`, `\*\_unix.go`, `\_ignoreme.go`.

#### Code style

Don't use line breaks within function call (on commas), try and fit one
function call on one line. No need for line wrapping of go code - one line is
one idea; a function call is one idea. If for whatever reasons it really does
feel too long, break it up in a different way, e.g. assign parameters to
variables first. Occasional lines with 120 characters are fine. Map and array
literals can be split over multiple lines, preferably only in assignments or
inferred declarations.

#### Testing

Ideally `TestSomething` is a self-contained function, so that tests can be
read without much jumping around - it's ok if these functions get big, they
are usually quite linear. Helpers with `t.Helper()` don't report where the
error happened so they should conceptually be a single assertion and also
called that e.g.

```go
func assertNumValue(t *testing.T, want float64, got Value) {
  t.Helper()
  // ...
}
```

[Effective Go]: https://go.dev/doc/effective_go
[Go Code Review Comments]: https://go.dev/wiki/CodeReviewComments
