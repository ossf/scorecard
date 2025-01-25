# `multi-scorecard`

This program runs OpenSSF Scorecard over many repositories using a [GitHub App](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps) credential.
GitHub is queried to determine the orgs and repos the app is installed on to determine which repos to run Scorecard over.

Results are printed to stdout in a JSON array.

*`multi-scorecard` was originally featured as part of [Jeff Mendoza](https://github.com/jeffmendoza) and [Stephen Augustus](https://github.com/justaugustus)' SOSS Fusion talk, "Scorecard at Scale: Old and New Possibilities for Lifting Security on All Repositories".*

- [Session page with slides](https://sched.co/1hcPq)
- [Session recording](https://youtu.be/-XZqbO3hGcw?si=eGicz0sjgiIRhol4)
- [Previous source repository](https://github.com/jeffmendoza/multi-scorecard)

## Usage

A [GitHub App](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps) must be created and installed on the repositories you wish to scan.

To install:

```console
go get github.com/ossf/scorecard/cmd/multi-scorecard@multi-scorecard
```

To run:

```console
multi-scorecard -appid 1234 -keyfile my-app.private-key.pem > results.json
```

Where `1234` is the App ID of the app, and `my-app.private-key.pem` is the private key file of the app.

Once the program has finished running, move the `results.json` file to the `src/results.json` in the scorecard-visualizer repo.

Example repo for local visualizer: http://localhost:3000/scorecard-visualizer/#/projects/github.com/uwu-tools/peribolos

## TODO

- Document required permissions for GitHub App
  - Contents: read-only
  - Metadata: read-only
- Add logging
- Resolve [workarounds](#workarounds)

### Workarounds

#### JSON format problems

```log
TS2345: Argument of type '(element: ScoreElement) => JSX.Element' is not assignable to parameter of type '(value: { details: null; score: number; reason: string; name: string; documentation: { url: string; short: string; }; }, index: number, array: { details: null; score: number; reason: string; name: string; documentation: { ...; }; }[]) => Element'.
  Types of parameters 'element' and 'value' are incompatible.
    Type '{ details: null; score: number; reason: string; name: string; documentation: { url: string; short: string; }; }' is not assignable to type 'ScoreElement'.
      Types of property 'details' are incompatible.
        Type 'null' is not assignable to type 'string[]'.
```

<details>

```log
    101 |
    102 |       <hr />
  > 103 |       {data.checks.map((element: ScoreElement) => (
        |                        ^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  > 104 |         <>
        | ^^^^^^^^^^
  > 105 |           <div key={element.name} className="card__wrapper">
        | ^^^^^^^^^^
  > 106 |             <div className="heading__wrapper" data-testid={element.name}>
        | ^^^^^^^^^^
  > 107 |               <h3>{element.name}</h3>
        | ^^^^^^^^^^
  > 108 |               {element.score !== -1 ? (
        | ^^^^^^^^^^
  > 109 |                 <span>{element.score}/10</span>
        | ^^^^^^^^^^
  > 110 |               ) : (
        | ^^^^^^^^^^
  > 111 |                 <NoAvailableDataMark />
        | ^^^^^^^^^^
  > 112 |               )}
        | ^^^^^^^^^^
  > 113 |             </div>
        | ^^^^^^^^^^
  > 114 |             <p>
        | ^^^^^^^^^^
  > 115 |               Description: {element.documentation.short.toLocaleLowerCase()}{" "}
        | ^^^^^^^^^^
  > 116 |               <a
        | ^^^^^^^^^^
  > 117 |                 href={`${element.documentation.url}`}
        | ^^^^^^^^^^
  > 118 |                 target="_blank"
        | ^^^^^^^^^^
  > 119 |                 rel="noreferrer"
        | ^^^^^^^^^^
  > 120 |               >
        | ^^^^^^^^^^
  > 121 |                 See documentation
        | ^^^^^^^^^^
  > 122 |               </a>
        | ^^^^^^^^^^
  > 123 |             </p>
        | ^^^^^^^^^^
  > 124 |             <p>Reasoning: {element?.reason.toLocaleLowerCase()}</p>
        | ^^^^^^^^^^
  > 125 |             {Array.isArray(element.details) && (
        | ^^^^^^^^^^
  > 126 |               <Collapsible details={element.details} />
        | ^^^^^^^^^^
  > 127 |             )}
        | ^^^^^^^^^^
  > 128 |           </div>
        | ^^^^^^^^^^
  > 129 |           <hr />
        | ^^^^^^^^^^
  > 130 |         </>
        | ^^^^^^^^^^
  > 131 |       ))}
        | ^^^^^^^^
    132 |     </>
    133 |   );
    134 | }
```

</details>

Replace all instances of:

```json
"details": null
```

with:

```json
"details": ["details", "string", "array"]
```

This could be a permissions issue with the GitHub App i.e., it does not have the requisite scope to read details of the check.
