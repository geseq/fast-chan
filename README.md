# fast-chan

Based on [Gringo](https://github.com/textnode/gringo) and compatible with [Gotemplate](https://github.com/ncw/gotemplate)

Benchmarks work best on GOMAXPROCS=1. In real world usage, this might have to be experimented with

This is created and optimized for SPSC case and might blow up for multiple producers and/or multiple consumers
