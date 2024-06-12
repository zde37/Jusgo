# Jusgo
Warning: Excessive use of this API may lead to uncontrollable laughter and increased productivity. (Use responsibly)!
# Description
Jusgo is a Programmer Joke API built with Golang, Clean Architecture and MongoDB. It provides a dedicated source of jokes tailored to the developer community and it's designed for seamless integration into developer tools and applications.
#
Our jokes are always up-to-date, even if your code isn't. We promise they won't throw any exceptions :)
# 
90% of Jusgo was built using the Go standard library. In some cases, all you need is the standard library.

## Available Endpoints

&#10004; Get multiple Jokes(paginated, default: 10)

&#10004; Get single Joke(by id)

&#10004; Add a Joke(Admin only)

&#10004; Update a Joke(Admin only)

&#10004; Delete a Joke(Admin only)

## How To Use
Live URL(Coming soon)

## Side Notes
You'll see this syntax alot in the code (Blasphemy!!!). Well...Http handlers should return errors.
```sh
func functionName(w http.ResponseWriter, r *http.Request) error {

}
```
#
Send me a PR if you know a good joke :)
##
