# Kuilei

This project is inspired by [Prow](https://github.com/kubernetes/test-infra/tree/master/prow), and implements chat-ops, automatic PR merging in a form of [Probot](https://github.com/probot/probot) Github App
> Prow is a Kubernetes based CI/CD system. Jobs can be triggered by various types of events and report their status to many different services. In addition to job execution, Prow provides GitHub automation in the form of policy enforcement, chat-ops via /foo style commands, and automatic PR merging.


## About the name Kuilei
> Kuilei(傀儡) is a kind of wooden puppet in ancient China, usually controlled from above using wires or strings depending on regional variations, and is often used in puppetry. 
> 
> There are also some masters who are proficient in mechanics who can make puppets that can move automatically, similar to robots today. The two most famous of them are [Mo Zi](https://en.wikipedia.org/wiki/Mozi) and [Lu Ban](https://en.wikipedia.org/wiki/Lu_Ban), who lived in the early portion of the Warring States period in Chinese history.

## Features

Commands:
- `/lgtm`: add `lgtm` label
- `/approve`: add `approved` label
- `/hold`: add `do-not-merge/hold` label
- `/kind bug`: add `kind/bug` label

Automatic PR merging:
- Merge the PR when label `lgtm` and `approved` both exist
- Do not merge the PR when label `do-not-merge/hold` exist

## Setup

```sh
# Install dependencies
npm install

# Run the bot
npm start
```

## Docker

```sh
# 1. Build container
docker build -t kuilei .

# 2. Start container
docker run -e APP_ID=<app-id> -e PRIVATE_KEY=<pem-value> kuilei
```

## Contributing

If you have suggestions for how kuilei could be improved, or want to report a bug, open an issue! We'd love all and any contributions.

For more, check out the [Contributing Guide](CONTRIBUTING.md).

## License

[ISC](LICENSE) © 2022 Airconduct
