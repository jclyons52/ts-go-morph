Test name: `TestSanitizedStackTraceDefeatsVSCodeGenericSecretRegex`

# Unsanitized input:

````
goroutine 7 [running]:
runtime/debug.Stack()
	runtime/debug/stack.go:26 +0x5e
github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls.(*LanguageService).getSignatureHelp(0x1)
	github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls/signature.go:42 +0x10
github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls.LookupKey(0x2)
	github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls/keys.go:7 +0x10
github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls.validateToken(0x3)
	github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls/token.go:9 +0x10
github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls.signRequest(0x4)
	github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls/sig.go:11 +0x10
github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls.setPwd(0x5)
	github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls/pwd.go:13 +0x10
````

# Sanitized output:

````
(REDACTED FRAME)
	(REDACTED FRAME)
(REDACTED FRAME)
	(REDACTED FRAME)
(REDACTED FRAME)
	(REDACTED FRAME)
(REDACTED FRAME)
	(REDACTED FRAME)
(REDACTED FRAME)
	(REDACTED FRAME)
(REDACTED FRAME)
	(REDACTED FRAME)
````
