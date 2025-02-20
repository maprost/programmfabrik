# programmfabrik

you need `exiftool` installed to run the application.

afterwards you can run it

```
go run main.go
```

You can test it with the following link in your browser: 

to see all tags:

http://localhost:8080/tags

for a special tag: 
take the tag's path like: `AAC::Main:ProfileType` and use it instead of `/tags`

http://localhost:8080/AAC::Main:ProfileType

if you interrupt the website, the server will also stop to run `exiftool`