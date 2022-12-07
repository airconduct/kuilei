// You can import your modules
// import index from '../src/index'

import nock from "nock";
// Requiring our app implementation
import myProbotApp from "../src";
import { Probot, ProbotOctokit } from "probot";
// Requiring our fixtures
import issues_opened_payload from "./fixtures/issues.opened.json";
import installation_created_payload from "./fixtures/installation_created.json"
import pull_request_opened_payload from "./fixtures/pull_request_opened.json"
import issue_comment_created_payload from "./fixtures/issue_comment_created.json"

const issueCreatedBody = { body: "Thanks for opening this issue!" };
const fs = require("fs");
const path = require("path");

const privateKey = fs.readFileSync(
  path.join(__dirname, "fixtures/mock-cert.pem"),
  "utf-8"
);

describe("My Probot app", () => {
  let probot: any;

  beforeEach(() => {
    nock.disableNetConnect();
    probot = new Probot({
      appId: 123,
      privateKey,
      // disable request throttling and retries for testing
      Octokit: ProbotOctokit.defaults({
        retry: { enabled: false },
        throttle: { enabled: false },
      }),
    });
    // Load our app into probot
    probot.load(myProbotApp);
  });

  test("create labels when installation created", async () => {
    const mock = nock("https://api.github.com")
      // Test that we correctly return a test token
      .post("/app/installations/2/access_tokens")
      .reply(200, {
        token: "test",
        permissions: {
          issues: "write",
        },
      })
      // Test list label
      .get("/repos/hiimbex/testing-things/labels")
      .reply(200, [])
      // Test create lgtm label
      .post("/repos/hiimbex/testing-things/labels", (body: any) => {
        expect(body).toMatchObject({
          name: "lgtm",
          color: "15DD18",
          description: "Indicates that a PR is ready to be merged.",
        });
        return true;
      })
      .reply(201)
      // Test create approved label
      .post("/repos/hiimbex/testing-things/labels", (body: any) => {
        expect(body).toMatchObject({
          name: "approved",
          color: "0FFA16",
          description: "Indicates a PR has been approved by an approver from all required OWNERS files.",
        });
        return true
      })
      .reply(201)
      // Test create kind/bug label
      .post("/repos/hiimbex/testing-things/labels", (body: any) => {
        expect(body).toMatchObject({
          name: "kind/bug",
          color: "D73A4A",
          description: "Categorizes issue or PR as related to a bug.",
        });
        return true
      })
      .reply(201)
      // Test create do-not-merge/hold label
      .post("/repos/hiimbex/testing-things/labels", (body: any) => {
        expect(body).toMatchObject({
          name: "do-not-merge/hold",
          color: "D73A4A",
          description: "Indicates that a PR should not merge because someone has issued a /hold command.",
        });
        return true
      })
      .reply(201)

    // Receive a webhook event
    await probot.receive({ name: "installation", payload: installation_created_payload });

    expect(mock.pendingMocks()).toStrictEqual([]);
  });

  test("creates a comment when an issue is opened", async () => {
    const mock = nock("https://api.github.com")
      // Test that we correctly return a test token
      .post("/app/installations/2/access_tokens")
      .reply(200, {
        token: "test",
        permissions: {
          issues: "write",
        },
      })

      // Test that a comment is posted
      .post("/repos/hiimbex/testing-things/issues/1/comments", (body: any) => {
        expect(body).toMatchObject(issueCreatedBody);
        return true;
      })
      .reply(200);

    // Receive a webhook event
    await probot.receive({ name: "issues", payload: issues_opened_payload });

    expect(mock.pendingMocks()).toStrictEqual([]);
  });

  test("when an pull request is opened, create check-runs and create comment", async ()=>{
    const mock = nock("https://api.github.com")
      // Test that we correctly return a test token
      .post("/app/installations/2/access_tokens")
      .reply(200, {
        token: "test",
        permissions: {
          issues: "write",
        },
      })
      // Get checkruns
      .get("/repos/hiimbex/testing-things/commits/d81102ea8baa41ec1d26a11db96aec1de887bffd/check-runs")
      .reply(200, {check_runs: []})
      // Create checkruns
      .post("/repos/hiimbex/testing-things/check-runs", (body:any)=>{
        expect(body).toMatchObject({
          name: "tide",
          head_sha: "d81102ea8baa41ec1d26a11db96aec1de887bffd",
          status: "in_progress"
        })
        return true
      })
      .reply(201);

    // Receive a webhook event
    await probot.receive({ name: "pull_request", payload: pull_request_opened_payload });
    expect(mock.pendingMocks()).toStrictEqual([]);
  })

  test("only create a comment when an pull request is opened", async ()=>{
    const mock = nock("https://api.github.com")
      // Test that we correctly return a test token
      .post("/app/installations/2/access_tokens")
      .reply(200, {
        token: "test",
        permissions: {
          issues: "write",
        },
      })
      // Get checkruns
      .get("/repos/hiimbex/testing-things/commits/d81102ea8baa41ec1d26a11db96aec1de887bffd/check-runs")
      .reply(200, {check_runs: [{name:"tide", id: 1}]})
      .patch("/repos/hiimbex/testing-things/check-runs/1", (body: any)=>{
        expect(body).toMatchObject({
          status: "in_progress"
        })
        return true
      })
      .reply(200)
      // Create checkruns
      .post("/repos/hiimbex/testing-things/check-runs", (body:any)=>{
        expect(body).toMatchObject({
          name: "tide",
          head_sha: "d81102ea8baa41ec1d26a11db96aec1de887bffd",
          status: "in_progress"
        })
        return true
      })
      .reply(200);

    // Receive a webhook event
    await probot.receive({ name: "pull_request", payload: pull_request_opened_payload });
    expect(mock.pendingMocks()).toStrictEqual(["POST https://api.github.com:443/repos/hiimbex/testing-things/check-runs"]);
  })

  test("add label when a comment is created", async ()=> {
    const mock = nock("https://api.github.com")
      // Test that we correctly return a test token
      .post("/app/installations/2/access_tokens")
      .reply(200, {
        token: "test",
        permissions: {
          issues: "write",
        },
      })
      // Mock add label api
      .post("/repos/hiimbex/testing-things/issues/3/labels", (body:any)=>{
        expect(body).toMatchObject({
          labels: ["lgtm", "approved", "kind/bug", "do-not-merge/hold"]
        })
        return true
      })
      .reply(200)
      // Mock owners file
      .get("/repos/hiimbex/testing-things/contents/OWNERS")
      .reply(200, "approvers:\n- lubingtan")

    // Receive a webhook event
    await probot.receive({ name: "issue_comment", payload: issue_comment_created_payload });
    expect(mock.pendingMocks()).toStrictEqual([]);
  })

  afterEach(() => {
    nock.cleanAll();
    nock.enableNetConnect();
  });
});

// For more information about testing with Jest see:
// https://facebook.github.io/jest/

// For more information about using TypeScript in your tests, Jest recommends:
// https://github.com/kulshekhar/ts-jest

// For more information about testing with Nock see:
// https://github.com/nock/nock
