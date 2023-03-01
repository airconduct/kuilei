<h1>Kuilei</h1>

This project is inspired by [Prow](https://github.com/kubernetes/test-infra/tree/master/prow), and implements chat-ops, automatic PR merging in a form of [Probot](https://github.com/probot/probot) Github App
> Prow is a Kubernetes based CI/CD system. Jobs can be triggered by various types of events and report their status to many different services. In addition to job execution, Prow provides GitHub automation in the form of policy enforcement, chat-ops via /foo style commands, and automatic PR merging.

- [About](#about)
- [Features](#features)
  - [Chat-Ops](#chat-ops)
  - [CI/CD Pipeline](#cicd-pipeline)
  - [Automatic Pull Requests Merging](#automatic-pull-requests-merging)
  - [Automatic notification](#automatic-notification)
  - [Mulitple Git Server Backend](#mulitple-git-server-backend)
- [Quick start](#quick-start)
  - [Create GitHub App](#create-github-app)
  - [Using docker to start hook server](#using-docker-to-start-hook-server)
- [Build From Source](#build-from-source)
- [Contributing](#contributing)
- [License](#license)

## About
> Kuilei(傀儡) is a kind of wooden puppet in ancient China, usually controlled from above using wires or strings depending on regional variations, and is often used in puppetry. 
> 
> There are also some masters who are proficient in mechanics who can make puppets that can move automatically, similar to robots today. The two most famous of them are [Mo Zi](https://en.wikipedia.org/wiki/Mozi) and [Lu Ban](https://en.wikipedia.org/wiki/Lu_Ban), who lived in the early portion of the Warring States period in Chinese history.

## Features

### Chat-Ops
Slash commands (commands with `/` as prefix):
- **Labels**
  - [x] `/lgtm [cancel]`: Add/Remove `lgtm` label to a pull request or issue.
  - [x] `/approve [cancel]`: Add/Remove `approved` label to a pull request or issue.
  - [x] `/[remove-]label xxx`: Add/Remove arbitrary label to a pull request or issue.
  - [ ] `/hold [cancel]`: Add/Remove `do-not-merge/hold` label to a pull request or issue.
- **Issue/PR management**
  - [ ] `/assign` Assign
  - [ ] `/close` Close
  - [ ] `/milestone`

Natural language parsing:
> *TODO*

### CI/CD Pipeline
Supported pipeline configuration:
- [ ] GitHub style pipeline configuration

Supported pipeline backend:
- [ ] Tekton pipeline
- [ ] Kubeflow pipeline (MLOps specific)
- [ ] Argo workflow
- [ ] Raw kubernetes
- [ ] Raw docker

### Automatic Pull Requests Merging
Just like `tide` in prow:
- [x] Merge pull requests with required labels
  - e.g. merge pull requests with labels `lgtm` and `approved`
- [x] Keep pull requests with certain labels from merging
  - e.g. hold pull requests with label `do-not-merge/hold`
- [ ] Reset mandatory CI job before merging
- [ ] Using merge-pool to manage multiple pull requests

### Automatic notification

- [ ] E-mail notification
- [ ] Slack notification
- [ ] WeCom notification
- [ ] DingTalk notification
- [ ] WeChat notification

### Mulitple Git Server Backend

- [x] GitHub
- [ ] GitLab
- [ ] Raw SSH Git Server
- [ ] Gerrit

## Quick start
### Create GitHub App
Follow the [official document](https://docs.github.com/en/apps/creating-github-apps/creating-github-apps/creating-a-github-app) to create your GitHub App.

Then go to the App settings page
- Get the `App ID`
- Set in the `webhook secret` blank and save it in local file `.env/web_secret`
- Generate a `priviate key`  and save it in local file `.env/private.pem`
- Set the `Webhook URL` to `http://<IP>:7771/hook`
### Using docker to start hook server

```sh
# Set App ID
APP_ID=
# Pull latest docker image
docker run -d --name kuilei --restart=always \
  -v `pwd`/.env:/etc/kuilei \
  -p 7771:7771 airconduct/kuilei hook \
  --github.appid ${APP_ID} \
  --github.hmac-token-file /etc/kuilei/web_secret \
  --github.private-key-file /etc/kuilei/private.pem \
  --address="0.0.0.0"
```


## Build From Source
```sh
# Build docker image
docker build -t kuilei .
```

## Contributing

If you have suggestions for how kuilei could be improved, or want to report a bug, open an issue! We'd love all and any contributions.

For more, check out the [Contributing Guide](CONTRIBUTING.md).

## License

[Apache](LICENSE) © 2022 Airconduct
