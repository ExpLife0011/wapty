# Roadmap 
## Implemented Features

* [x] Proxy
	* [x] Backend
	* [x] UI
* [x] History
	* [x] Backend
	* [x] UI
* [ ] Repeat
	* [x] Backend
	* [ ] UI
* [ ] Decode 
	* [x] Backend
	* [x] CLI
	* [ ] UI
	* Codecs
		* [ ] Binary
		* [x] Base 16 
		* [ ] Base32
		* [x] Base64
		* [ ] URL
		* [ ] HTML Entities
		* [ ] Javascript-Escape
		* [ ] Gzip
* [ ] Save/Load
* [ ] Intrude
* [ ] Sequence
* [ ] Compare
* [ ] Crawl
* [ ] Scan
* [ ] Mock
* [ ] User Documentation
* [ ] Extend
* [ ] Dashboard

## Packages and Coverage
Package Name | Currently Actively Developing           | Development Phase           | Test Coverage
:-----------:|:---------------------------------------:|:---------------------------:|:--:
main         | [Empijei](https://github.com/empijei)   | Usable                      | [![main](https://gocover.io/_badge/github.com/empijei/wapty)](http://gocover.io/github.com/empijei/wapty)
mitm         | [Empijei](https://github.com/empijei)   | Done                        | [![mitm](https://gocover.io/_badge/github.com/empijei/wapty/mitm)](http://gocover.io/github.com/empijei/wapty/mitm)
intercept    | [Empijei](https://github.com/empijei)   | Usable                      | [![intercept](https://gocover.io/_badge/github.com/empijei/wapty/intercept)](http://gocover.io/github.com/empijei/wapty/intercept)
decode       | [AnnaOpss](https://github.com/annaopss) | Usable, missing some codecs | [![decode](https://gocover.io/_badge/github.com/empijei/wapty/decode)](http://gocover.io/github.com/empijei/wapty/decode)
repeat       | [Empijei](https://github.com/empijei)   | Only backend implemented    | [![repeat](https://gocover.io/_badge/github.com/empijei/wapty/repeat)](http://gocover.io/github.com/empijei/wapty/repeat)
ui           | [Empijei](https://github.com/empijei)   | Usable                      | [![ui](https://gocover.io/_badge/github.com/empijei/wapty/ui)](http://gocover.io/github.com/empijei/wapty/ui)
sequence     | ✗                                       | Draft                       | N/A
compare      | ✗                                       | Draft                       | N/A
plug         | ✗                                       | Not implemented yet         | N/A
mocksy       | ✗                                       | Only Burp importer written  | N/A

# Detailed TODOs
## Initial stage 
This stage will be the first stage for wapty, before this is finished wapty will likely have unstable APIs and won't be really usable.

* [x] Implement Proxy
* [x] Implement History
* [x] Use a formal approach to fuzzy decoding
* [x] Refactor Decode package
* [x] Rewrite UI in gopherjs
* [x] Simplify server-side code for UI
* [ ] finish Repeat tests
* [x] use templates for UI
* [ ] Add UI to repeater
* [x] Use https://gocover.io to compute coverage
* [ ] Add saving functionality
* [ ] Add scoping
* [ ] Add history filtering/sorting
* [x] Add functionality: releasing the intercept should forward all pending requests
* [ ] Add internal router
* [ ] ignore recursive connect
* [ ] Add intercept checker in the right spots
* [ ] Allow the user to change the destination endpoint
* [ ] Serve the certificates on a specific fake host/path
* [ ] Add req ID to editor and reject unexpected requests
* [ ] Add Intruder
* [ ] Allow for creating multiple proxies, change ports. 
* [ ] Keep track of which proxy intercepted the request in metadata.
* [ ] [UI] Send the whole status on ui connect
* [ ] [UI] Sanitize metadata
* [ ] [UI] show already pending request/response upon connection
* [ ] [UI] error log
* [ ] [UI] auto-open ui in browser on launch
* [ ] [UI] monospace textareas
* [ ] [UI] resizable splits
* [ ] Look for fixmes and todos in the code
* [ ] Provide a ui to the decoder

The following is just some general polishing before calling this a proper project
* [ ] Improve README
* [ ] Handle panics within the package
* [ ] Move all constant strings to actual constants
* [ ] All the deferred closes if err!= nil send that, otherwise propagate the new one
* [ ] Doc comment should be a complete sentence that starts with the name being declared.
* [ ] Lint the code, improve score on goreportcard
* [ ] general code polish, doc and and testing

## Moving to Release
This is meant to be mostly an improvement, adding features that are less used in burpsuite but are still there and should end up in wapty before it is called a proper replacement for burp

* [ ] Add AutoEdit
* [ ] Add cURL converter
* [ ] Default to bare sockets on error
* [ ] profile the code, try to find limit-cases
* [ ] Add Spider (remember to add timeouts §8.10)
* [ ] Add Scanner
* [ ] Add Sequencer
* [ ] Add recursive intruder with flows
* [ ] Add syntax highlight for relevant buffers
* [ ] Test transparent proxying
* [ ] Allow to transparently remap a local port to another one with custom certificate. see [tlsmitm](https://github.com/empijei/tlsmitm) as a reference

## Release
* [ ] Have penetration testers use wapty for a while, collect feedback
* [ ] Implement fixes, add suggestions to a feature list
* [ ] Advertise and publish the project on a broader scale

## Improvements
This section contains the features that burpsuite lacks but that will make this project different :)

These features will probably be implemented along with the ones in the other stages.

* [ ] Add Mocksy
* [ ] UI add preview of blocked requests queue with a chance to perform some actions on them
* [ ] Add pre-engagement 
	* [ ] analysis/recon, 
	* [ ] detect technologies used/versions
* [ ] Deserializer of java/flash/php serialized objects (maybe editor?)
* [ ] Add SAML, JWT decoder/editor
* [ ] Add a Pathfinder feature to spider that allows to backtrace how a certain URL was discovered
* [ ] Add a Plugin manager / Make plugin behave as package testing, just plug the stuff
* [ ] Add a SQLmap invoker
* [ ] Add fuzzing payloads generator
* [ ] Add TUI
* [ ] Add scripting engine (JS/Lua)

## Misc:
These are the feature I still don't know if are worth adding

(PRs are welcome :D )

* [ ] Add Content-Length override
* [ ] Add Beautifier
* [ ] Decompress HTTP2 instead of disabling it
* [ ] [UI] Make operations unblocking and detect ui freezes/deaths. If channel is full and not being read, kill the client.