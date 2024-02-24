# Evy Source Philosophy

Our goal is to keep Evy's codebase as clear and understandable as the language
itself. Evy is designed to be a learning experience in every way, and that
includes the way the code itself is written. Our main goal is to make
programming accessible and inviting, a philosophy we apply to the project's
code too.

Wherever possible, we avoid using frameworks or extra tools, preferring
a "from scratch" approach that emphasizes learning. This means if you explore
Evy's website using your browser's "View Page Source" feature, the code you
see (HTML, CSS, and JavaScript) is what we wrote – there's no hidden build
process changing things. For more on this, see Julia Evans'
[Writing Javascript without a build system] and David Heinemeier Hansson's
[\[...\] nobuild for the front-end]

The same applies to our documentation. Whether it's the repository's
[README.md], the guides on [docs.evy.dev], or the automatically generated [godocs],
everything should be well-organized and easy to understand.

We favor clear, readable code over complex shortcuts. For example, in
JavaScript, we often use for-loops instead of dense map-reduce-filter chains
when it makes the code easier to understand. Our Go code strives to follow
the best practices outlined [Go Code Review Comments] and [Effective Go].

Finally, we put care into crafting clear commit messages and pull requests
following our contributing guidelines. This helps everyone track the
project's evolution and track down issues.

[\[...\] nobuild for the front-end]: https://world.hey.com/dhh/once-1-is-entirely-nobuild-for-the-front-end-ce56f6d7
[docs.evy.dev]: https://docs.evy.dev
[Effective Go]: https://go.dev/doc/effective_go
[Go Code Review Comments]: https://go.dev/wiki/CodeReviewComments
[README.md]: https://github.com/evylang/evy/blob/main/README.md
[Writing Javascript without a build system]: https://jvns.ca/blog/2023/02/16/writing-javascript-without-a-build-system/
