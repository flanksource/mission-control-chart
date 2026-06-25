// Branch strategy is decided at runtime from the triggering event, because the
// shared flanksource/action-workflows create-release.yml takes no inputs and
// reads only this config file:
//   - push to main      -> beta pre-release (X.Y.Z-beta.N)
//   - workflow_dispatch  -> stable release  (X.Y.Z)
//
// `dummy-release` is a branch that never exists. semantic-release refuses to
// run a pre-release branch unless at least one stable branch is also configured,
// so it satisfies that rule without ever cutting a stable release.
const isManual = process.env.GITHUB_EVENT_NAME === "workflow_dispatch";

module.exports = {
  branches: isManual
    ? ["main"]
    : [{ name: "main", channel: "beta", prerelease: "beta" }, { name: "dummy-release" }],
  plugins: [
    [
      "@semantic-release/commit-analyzer",
      {
        releaseRules: [
          { type: "doc", scope: "README", release: "patch" },
          { type: "fix", release: "patch" },
          { type: "chore", release: "patch" },
          { type: "refactor", release: "patch" },
          { type: "feat", release: "patch" },
          { type: "ci", release: "patch" },
          { type: "style", release: "patch" },
        ],
        parserOpts: {
          noteKeywords: ["MAJOR RELEASE"],
        },
      },
    ],
    "@semantic-release/release-notes-generator",
    [
      // From: https://github.com/semantic-release/github/pull/487#issuecomment-1486298997
      "@semantic-release/github",
      {
        successComment: false,
        failTitle: false,
      },
    ],
  ],
};
