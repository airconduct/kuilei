import { Probot } from "probot";
import yaml from "js-yaml";

// const yaml_load = <T = ReturnType<typeof original>>(...args: Parameters<typeof original>): T => yaml_load(...args);


const needLabels = [
  {
    name: "lgtm",
    color: "15DD18",
    description: "Indicates that a PR is ready to be merged.",
  },
  {
    name: "approved",
    color: "0FFA16",
    description: "Indicates a PR has been approved by an approver from all required OWNERS files.",
  },
  {
    name: "kind/bug",
    color: "D73A4A",
    description: "Categorizes issue or PR as related to a bug.",
  },
  {
    name: "do-not-merge/hold",
    color: "D73A4A",
    description: "Indicates that a PR should not merge because someone has issued a /hold command.",
  },
]

const commandLabels = new Map([
  ["lgtm", "lgtm"],
  ["approve", "approved"],
  ["kind/bug", "kind/bug"],
  ["hold", "do-not-merge/hold"],
])

const tideLabelsMustExist = new Map([
  ["lgtm", false],
  ["approved", false],
])

const tideLabelsMustNotExist = new Map([
  ["do-not-merge/hold", false],
])

export = (app: Probot) => {
  app.on("installation.created", async (context) => {
    const owner = context.payload.installation.account.login

    if (context.payload.repositories) {
      await Promise.all(context.payload.repositories.map(async (repo) => {
        const labels = await context.octokit.issues.listLabelsForRepo({
          owner: owner,
          repo: repo.name
        })
        var currentLabels = new Map()
        labels.data.forEach((label)=>{
          currentLabels.set(label.name, label)
        })
        await Promise.all(needLabels.map(async (label)=>{
          if (!currentLabels.has(label.name)) {
            await context.octokit.issues.createLabel({
              owner: owner,
              repo: repo.name,
              name: label.name,
              color: label.color,
              description: label.description
            })
          }
        }))
      }));
    }
  })
  app.on("installation_repositories.added", async (context)=>{
    const owner = context.payload.installation.account.login
    await Promise.all(context.payload.repositories_added.map(async (repo) => {
      await Promise.all(needLabels.map(async (label)=>{
        await context.octokit.issues.createLabel({
          owner: owner,
          repo: repo.name,
          name: label.name,
          color: label.color,
          description: label.description
        })
      }))
    }));
  })

  app.on("issues.opened", async (context) => {
    await context.octokit.issues.createComment(context.issue({
      body: "Thanks for opening this issue!",
    }));
  });

  app.on(["pull_request.opened", "pull_request.edited", "pull_request.reopened"], async (context) => {
    // Create comment
    context.payload.pull_request.head.ref
    await context.octokit.issues.createComment(context.issue({
      body: "Thanks for contributing!"
    }))
    // Get tide check
    const resp = await context.octokit.checks.listForRef(context.repo({ref: context.payload.pull_request.head.ref}))
    if (resp.data.check_runs) {
      for (let i=0;i<resp.data.check_runs.length;i++) {
        const check_run  = resp.data.check_runs.at(i)
        if (check_run && check_run.name == "tide") {
          // Exists
          // udpate tide check
          await context.octokit.checks.update(context.repo({
            check_run_id: check_run.id,
            status: "in_progress"
          }))
          return
        }
      }
    }
    // Create tide check
    await context.octokit.checks.create(context.pullRequest({
      name: "tide",
      head_sha: context.payload.pull_request.head.sha,
      status: "in_progress"
    }))
  });

  app.on(["issue_comment.created", "issue_comment.edited"], async (context) => {
    if (context.isBot) {
      return
    }
    if (!context.payload.issue.pull_request) {
      return
    }
    const resp = await context.octokit.request(context.repo({
      method: "GET",
      url: "/repos/{owner}/{repo}/contents/{path}",
      path: "OWNERS",
      mediaType: {
        format: "raw",
    }}))
    context.log({"content_resp____________is": resp})
    const owners_config = yaml.load(resp.data) as {approvers:string[],reviewers:string[]}
    const username = context.payload.comment.user.login.toLowerCase()
    const approveAuth = owners_config.approvers.includes(username)
    const lgtmAuth = approveAuth || owners_config.reviewers.includes(username)
    // TODO: support more commands
    const commands = getCommands(context.payload.comment.body)
    var labels : string[] = []
    commands.forEach((val)=>{
      if ( val && commandLabels.has(val)) {
        const label = commandLabels.get(val)
        if (label) {
          switch (label) {
            case "lgtm":
              lgtmAuth && labels.push(label);
              break;
            case "approved":
              approveAuth && labels.push(label)
              break;
            default:
              labels.push(label)
          }
        }
      }
    })
    await context.octokit.issues.addLabels(context.issue({
      labels: labels,
    }))
  });

  app.on(["pull_request.labeled", "pull_request.unlabeled"], async (context) => {
    var checkRunID : number | undefined = undefined
    // Get tide check_run_id
    const resp = await context.octokit.checks.listForRef(context.repo({ref: context.payload.pull_request.head.ref}))
    if (resp.data.check_runs) {
      for (let i=0;i<resp.data.check_runs.length;i++) {
        const checkRun = resp.data.check_runs.at(i)
        if (checkRun && checkRun.name == "tide") {
          checkRunID = checkRun.id
        }
      }
    }
    // Not Exist
    if (!checkRunID) {
      // Create tide check
      const resp = await context.octokit.checks.create(context.pullRequest({
        name: "tide",
        head_sha: context.payload.pull_request.head.sha,
        status: "in_progress"
      }))
      checkRunID = resp.data.id
    }

    const labels = context.payload.pull_request.labels
    var labelsMustExist = new Map(tideLabelsMustExist)
    var labelsMustNotExist = new Map(tideLabelsMustNotExist)
    labels.forEach((label)=>{
      if (tideLabelsMustExist.has(label.name)) {
        labelsMustExist.set(label.name, true)
      }
      if (labelsMustNotExist.has(label.name)) {
        labelsMustNotExist.set(label.name, true)
      }
    })

    var tidePass = true
    labelsMustExist.forEach((exist)=>{
      if (!exist) {
        tidePass = false
      }
    })
    labelsMustNotExist.forEach((exist)=>{
      if (exist) {
        tidePass = false
      }
    })
    if (!tidePass) {
      return
    }

    // Update status
    if (checkRunID) {
      const now =new Date()
      context.octokit.checks.update(context.repo({
        check_run_id: checkRunID,
        status: "completed",
        conclusion: "success",
        completed_at: now.toISOString()
      }))
    }
  })

  app.on("check_suite.completed", async (context)=>{
    const checkRuns = await context.octokit.checks.listForRef(context.repo({
      ref: context.payload.check_suite.head_sha
    }))
    var allCheckPass = true
    checkRuns.data.check_runs.forEach((checkRun)=>{
      if (checkRun.status!="completed") {
        allCheckPass = false
      }
    })

    if (allCheckPass) {
      const prs = context.payload.check_suite.pull_requests
      await Promise.all(prs.map(async (pr)=>{
        await context.octokit.pulls.merge(context.repo({
          pull_number: pr.number,
          merge_method: "rebase"
        }))
      }))
    }
  })
  // For more information on building apps:
  // https://probot.github.io/docs/

  // To get your app running against GitHub, see:
  // https://probot.github.io/docs/development/
};

function getCommands(txt: string) : string[] {
  var commands : string[] = []
  txt.split("\n").forEach((line)=>{
    const cmd = getCommand(line)
    if (cmd) {
      commands.push(cmd)
    }
  })
  return commands
}

function getCommand(lineRaw: string) : string | undefined {
  const line = lineRaw.replace(/(\r\n|\n|\r)/gm, "");
  if (line.length == 0) {
    return
  }
  if (line.at(0) != "/") {
    return
  }

  const cmdArg = line.substring(1)

  const parts = cmdArg.split(" ", 2)
  const cmd = parts.at(0)
  const arg = parts.at(1)
  if (cmd && arg) {
    if (cmd == "kind") {
      return cmd+"/"+arg
    }
  }
  return cmd
}
