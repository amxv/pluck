export const siteConfig = {
  name: "pluck",
  strapline: "Bring any GitHub folder or file home",
  description:
    "A small Go CLI for copying folders or files from GitHub URLs into a local directory.",
  repoUrl: "https://github.com/amxv/pluck",
  accentColor: "#b7f34a",
  accentColorDark: "#d7ff80",
  footerSections: [
    {
      title: "pluck",
      text:
        "A focused command for turning a GitHub folder or file URL into local files, with native binaries and a tiny npm wrapper."
    },
    {
      title: "What this site covers",
      text:
        "Installation, output paths, GitHub authentication, releases, and the docs site itself."
    },
    {
      title: "Repository",
      linkPrefix: "Source: ",
      linkHref: "https://github.com/amxv/pluck",
      linkLabel: "github.com/amxv/pluck"
    }
  ]
} as const;

export const docCategories = [
  "Start",
  "Usage",
  "Distribution",
  "Reference"
] as const;

export const primaryNav = [
  { href: "/docs", label: "Docs" },
  { href: siteConfig.repoUrl, label: "GitHub", external: true }
];
