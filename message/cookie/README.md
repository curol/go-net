# Cookie

HTTP cookies are small pieces of data that are sent from a web server to a client (usually a web browser) and stored on the client's computer.
Cookies are used to maintain state between requests in a stateless protocol like HTTP.

When a client sends an HTTP request to a server, it can include any cookies that were previously set by the server for that domain.
The server can then use the cookies to identify the client and **maintain state between requests**.

Cookies can have various attributes, such as a name, a value, an expiration time, a domain, a path, and a secure flag.
The name and value are used to store data, while the other attributes control how the cookie is transmitted and stored.

Cookies can be used for various purposes, such as session management, user tracking, and personalization.
However, cookies can also be used for tracking and advertising purposes, which has led to concerns about privacy and security.

Web browsers typically allow users to view and delete cookies, and some browsers also allow users to block cookies entirely or only accept cookies from certain domains.

1. Request Example:

```
GET / HTTP/1.1
Host: example.com
Cookie: mycookie=value; othercookie=othervalue
```

2. Response Example:

```
HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
Set-Cookie: mycookie=value; Expires=Wed, 21 Oct 2015 07:28:00 GMT; Path=/
Set-Cookie: othercookie=othervalue; Expires=Wed, 21 Oct 2015 07:28:00 GMT; Path=/
````

## Why pass the url for cookie

When setting cookies in an HTTP response, you can optionally specify a `Domain` attribute for the cookie. The `Domain` attribute specifies the domain for which the cookie is valid. If the `Domain` attribute is not specified, the cookie is only valid for the domain that set the cookie.

When setting a cookie in an HTTP response, you typically want the cookie to be valid for the domain that the response is sent to. To determine the domain, you can use the `Request.URL` field of the HTTP request that triggered the response. The `Request.URL` field contains the URL of the request, which includes the domain.

By passing the `Request.URL` field to the `SetCookies` method of the `http.CookieJar` interface, you allow the cookie jar to determine the domain for the cookie automatically. The cookie jar can then store the cookie in the appropriate storage location for that domain.

Note that if you want to set a cookie for a specific subdomain or path, you can also specify the `Domain` and `Path` attributes of the cookie explicitly.

## Domain
In the context of HTTP, a domain is a string that identifies a group of related resources on the Internet. A domain name is typically used to identify a web server that hosts a website or web application.

In HTTP, the domain is used to determine the scope of a cookie. When a server sets a cookie in an HTTP response, it can specify a `Domain` attribute for the cookie. The `Domain` attribute specifies the domain for which the cookie is valid. If the `Domain` attribute is not specified, the cookie is only valid for the domain that set the cookie.

For example, if a server sets a cookie with a `Domain` attribute of "example.com", the cookie is valid for all subdomains of "example.com", such as "www.example.com" and "blog.example.com". However, the cookie is not valid for other domains, such as "google.com" or "facebook.com".

Note that the domain name is part of the URL that a client uses to access a resource on the Internet. The domain name is typically resolved to an IP address using the Domain Name System (DNS) before the client can establish a connection to the server.