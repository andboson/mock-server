# Glossary

**Expectation**
One rule: a Matcher paired with a Mock response. Newest Expectation wins when several match.

**Matcher**
The part of an Expectation that decides whether an incoming request applies: method, path, body, and (planned) headers.

**Mock response**
What the server returns when an Expectation matches: status, headers and body.
_Wire names:_ `status`, `response_headers`, `mock` (the body).

**History entry**
One recorded incoming request, whether or not it matched an Expectation.

**Scenario**
A named set of Expectations that is loaded together to set up one test situation.

**Unmatched request**
A request that no Expectation matches. It gets a 404 and is still recorded as a History entry.
