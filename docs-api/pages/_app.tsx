import React from "react";
import Head from "next/head";
import { Lato } from "@next/font/google";

import "prismjs";
import "prismjs/components/prism-bash.min";
import "prismjs/components/prism-graphql.min";
import "prismjs/components/prism-javascript.min";
import "prismjs/components/prism-json.min";
import "prismjs/components/prism-jsx.min";
import "prismjs/components/prism-markup.min";
import "prismjs/components/prism-toml.min";
import "prismjs/components/prism-typescript.min";
import "prismjs/components/prism-yaml.min";
import "prismjs/plugins/autolinker/prism-autolinker.min";
import "prismjs/themes/prism.css";

import "../styles/globals.scss";

import type { AppProps } from "next/app";
import type { MarkdocNextJsPageProps } from "@markdoc/next.js";
import { TopNav } from "../components/TopNav";
import { TableOfContents } from "../components/TableOfContents";
import { ThemeProvider } from "next-themes";

const TITLE = "ZITADEL API reference";
const DESCRIPTION = "";

function collectHeadings(node, sections = []) {
  if (node) {
    if (node.name === "Heading") {
      const title = node.children[0];

      if (typeof title === "string") {
        sections.push({
          ...node.attributes,
          title,
        });
      }
    }

    if (node.children) {
      for (const child of node.children) {
        collectHeadings(child, sections);
      }
    }
  }

  return sections;
}

export type MyAppProps = MarkdocNextJsPageProps;

// const lato = Lato({ subsets: ["latin"], weight: ["400"] });

export default function MyApp({ Component, pageProps }: AppProps<MyAppProps>) {
  const { markdoc } = pageProps;

  let title = TITLE;
  let description = DESCRIPTION;

  if (markdoc) {
    if (markdoc.frontmatter.title) {
      title = markdoc.frontmatter.title;
    }
    if (markdoc.frontmatter.description) {
      description = markdoc.frontmatter.description;
    }
  }

  const toc = pageProps.markdoc?.content
    ? collectHeadings(pageProps.markdoc.content)
    : [];

  return (
    <>
      <Head>
        <title>{title}</title>
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <meta name="referrer" content="strict-origin" />
        <meta name="title" content={title} />
        <meta name="description" content={description} />
        <link rel="shortcut icon" href="/favicon.ico" />
        <link rel="icon" href="/favicon.ico" />
      </Head>

      <ThemeProvider
        attribute="class"
        defaultTheme="system"
        storageKey="cp-theme"
      >
        <div
          className={`bg-white dark:bg-background-dark-500 text-gray-800 dark:text-gray-100 `}
        >
          <div className="flex flex-grow max-w-8xl mx-auto relative">
            <TableOfContents toc={toc} />
            <div className="overflow-auto flex-grow pt-0 pl-4 lg:pl-10 pr-8 lg:pr-10 pb-8">
              <TopNav></TopNav>
              <main className="">
                <Component {...pageProps} />
              </main>
            </div>
          </div>
        </div>
      </ThemeProvider>
    </>
  );
}

// MyApp.getInitialProps = async (ctx) => {
//   //   const protoPath = "admin.proto";
//   //   const text = readFileSync(
//   //     join(__dirname, `../../proto/zitadel/${protoPath}`),
//   //     "utf8"
//   //   );

//   //   console.log(text);
//   console.log(ctx);
//   return {
//     ...ctx,
//     props: {},
//   };
// };
