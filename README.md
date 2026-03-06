# go-gitter

I'm implementing git, in Go, from scratch. The goal is to learn both the Go language by doing a relatively challenging project, and the git internals.

The plan, of course, is not to implement 100% of git. I intend to implement only a small subset of the main commands, and even those, in their most basic variations.

It should go without saying, but this is a toy project. Don't use it for anything serious.

## Status

Right now, the only thing I have implemented is the `init` command, and that only in its most basic form. Currently, this is the list of commands I intend to implement:

- add
- status
- commit
- log

I'm not sure whether I'll try to implement `merge` and the commands that involve networking. Most likely not, but let's see.

## License

MIT
