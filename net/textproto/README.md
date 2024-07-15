# Text Protocol

## Parsing and Serialization

The general term for parsing and serializing is "Data Marshalling" or just "Marshalling". 

This is an abstraction of Textproto for reading and writing sinple text messages over a connection.

- **Parsing** is the process of converting data in a specific format (like JSON, XML, etc.) into a format that your program can use, such as a data structure or object. This is also known as "unmarshalling" or "deserialization". For example, when a client sends a request to the server, convert the raw data or bytes to a request struct.

- **Serializing** is the process of converting a data structure or object in your program into a format that can be stored or transmitted, such as a JSON or XML string. This is also known as "marshalling". For example, converting a request struct to bytes.

Together, these processes are used to convert data between formats, often for the purposes of storage, transmission over a network, or communication between different parts of a program.

## Body

The body is not closed because the body of a message may be too big for memory. Instead its open and waiting to be read into storage or memory depending on where the server wants it.


## Domain

The `Domain` in the context of HTTP and web technologies refers to the part of a network address that identifies it as belonging to a particular domain. It is used to specify the internet realm of an entity or organization. Domains are structured in a hierarchical manner, with the top-level domain (TLD) at the highest level (e.g., `.com`, `.org`, `.net`, `.gov`, `.edu`), followed by the second-level domain (SLD), which is typically the name of the organization or entity (e.g., `example` in `example.com`), and potentially further subdivided into subdomains (e.g., `sub.example.com`).

In the context of the `Host` header from an HTTP/1.1 request, the domain name specifies the server's domain to which the request is being sent. This is crucial for the server to understand which website or web application the client intends to communicate with, especially in environments where a single server hosts multiple domains (virtual hosting). The `Host` header may include both the domain name and, optionally, the port number if the request is targeting a specific port other than the default (port 80 for HTTP and port 443 for HTTPS).

## Cookie

An HTTP cookie is a small piece of data sent from a website and stored on the user's computer by the user's web browser while the user is browsing. Cookies were designed to be a reliable mechanism for websites to remember stateful information (such as items added in the shopping cart in an online store) or to record the user's browsing activity (including clicking particular buttons, logging in, or recording which pages were visited in the past). They can also be used to remember arbitrary pieces of information that the user previously entered into form fields, such as names, addresses, passwords, and credit card numbers.

Cookies perform essential functions in the modern web. For example, they are used to keep a user logged into a website by remembering their login details. They can also be used for targeted advertising, to customize the user experience, or for analytics about how users interact with a website.

Cookies are set by a web server and sent to a user's browser, which stores them for later use. The next time the user visits the same server, the browser sends the cookies back to the server, allowing the server to recognize the user and respond appropriately.

### Cookie Attributes

Cookie attributes define various properties of a cookie to control its behavior in the client's browser. 

- `Path`: Specifies a URL path that must exist in the requested URL for the browser to send the Cookie header. (`Path=/` means the cookie is available throughout the entire domain.)
- `Expires`: Defines the lifetime of the cookie. The browser will delete the cookie after the specified date and time. (`Expires=Wed, 09 Jun 2021 10:18:14 GMT` sets a specific expiration date for the cookie.)
- `HttpOnly`: Marks the cookie as accessible only through the HTTP protocol. This means the cookie cannot be accessed through client-side scripts, which is a measure against cross-site scripting (XSS) attacks.
- `Secure`: Indicates that the cookie should only be sent over secure (HTTPS) connections. This helps to ensure the confidentiality and integrity of the transmitted information.

These attributes enhance the security and functionality of cookies, guiding the browser on how to handle them.

Example of a raw HTTP response that includes a `Set-Cookie` header with various cookie attributes, including the `SameSite` attribute:

```
HTTP/1.1 200 OK
Content-Type: text/html; charset=UTF-8
Set-Cookie: sessionId=38afes7a8; Path=/; Expires=Wed, 09 Jun 2021 10:18:14 GMT; HttpOnly; Secure; SameSite=Lax

<html>
<head>
    <title>Cookie Test</title>
</head>
<body>
    <p>Cookie with attributes set.</p>
</body>
</html>
```

In this response:

- `sessionId=38afes7a8` is the cookie being set.
- `Path=/` indicates that the cookie is available throughout the entire domain.
- `Expires=Wed, 09 Jun 2021 10:18:14 GMT` specifies when the cookie will expire.
- `HttpOnly` means the cookie is not accessible via JavaScript, enhancing security by preventing cross-site scripting (XSS) attacks.
- `Secure` indicates that the cookie should only be sent over secure (HTTPS) connections.
- `SameSite=Lax` controls how cookies are sent with cross-site requests. `Lax` allows the cookie to be sent with top-level navigations and will prevent sending cookies with cross-site subrequests (like images or frames).

### Sanesite attribute

 The `SameSite` attribute is used in HTTP `Set-Cookie` response headers, not in request headers. It instructs the browser on how to handle cookies across site requests to enhance security.

Here's an example of how a server might include the `SameSite` attribute in a `Set-Cookie` header in its HTTP response:

```
HTTP/1.1 200 OK
Content-Type: text/html
Set-Cookie: session_token=abc123; Path=/; HttpOnly; SameSite=Lax
Set-Cookie: theme=light; Path=/; Expires=Wed, 09 Jun 2021 10:18:14 GMT; SameSite=Strict

<html>
<body>
    <p>Example HTML content here.</p>
</body>
</html>
```

In this response:

- The `session_token` cookie is set with `SameSite=Lax`, allowing it to be sent in top-level navigations to the server from other sites.
- The `theme` cookie is set with `SameSite=Strict`, meaning it will only be sent in requests originating from the same site as the cookie.

This is how you would see the `SameSite` attribute being used in practice, in a server's HTTP response, rather than in a request from a client like a browser.

### Samesite cookie values

The `SameSite` cookie attribute accepts three values, each of which controls how cookies are sent with cross-site requests:

1. **`Strict`**: The cookie will only be sent with requests initiated from the same site the cookie belongs to. This is a good option for cookies that are strictly necessary for the functionality of your site's core features.

2. **`Lax`**: Cookies are not sent on normal cross-site subrequests (for example, loading images into a third party site), but are sent when a user is navigating to the origin site (i.e., when following a link). This is a reasonable balance between security and usability for cookies needed for features that are intended to be accessible across sites.

3. **`None`**: Cookies will be sent in all contexts, i.e., in responses to both first-party and cross-origin requests. If `SameSite=None` is set, the cookie `Secure` attribute must also be set (or the cookie will be blocked).

The `SameSite` attribute is used to mitigate the risk of cross-origin information leakage, and provides some protection against cross-site request forgery attacks (CSRF).

### Cookie Domain

The `Domain` attribute in a cookie specifies the domain for which the cookie is valid and can be sent by the browser in HTTP requests. This attribute is used to control the scope of the cookie: it determines which domains (including subdomains) can receive the cookie in HTTP requests originating from the browser.

- **Default Behavior**: If the `Domain` attribute is not specified, the cookie will only be sent to the domain that set the cookie, not including subdomains.

- **Specifying the Domain**: When the `Domain` attribute is set, the cookie will be sent to the specified domain and all its subdomains. For example, if a cookie is set with `Domain=example.com`, it will be sent with requests to `example.com`, `sub.example.com`, and `another.sub.example.com`.

- **Security Consideration**: Setting the `Domain` attribute to a broader domain (e.g., setting a cookie from `sub.example.com` with `Domain=example.com`) can make the cookie available to a larger set of subdomains, which might have security implications. It's important to set this attribute carefully to avoid unintentionally sharing session information across domains that should not share such information.

The `Domain` attribute is part of the cookie string that can be set in an HTTP `Set-Cookie` header from the server. Here's an example of how it might look in an HTTP response:

```
Set-Cookie: sessionId=abc123; Domain=example.com; Path=/; Secure; HttpOnly
```

In this example, the `sessionId` cookie will be available to `example.com` and any of its subdomains.

### Cookie and host header

In HTTP, both the `Cookie` and `Host` headers play significant roles in web communication, but they serve different purposes:

- **Cookie Header**: The `Cookie` header is sent by the web browser to the server in HTTP requests. It contains previously stored cookies from the server (set via `Set-Cookie` headers in previous responses). Cookies are used for maintaining session state, personalization, and tracking user behavior across visits and page requests. Each cookie has a name and a value, and optionally, attributes like `Expires`, `Max-Age`, `Domain`, `Path`, `Secure`, `HttpOnly`, and `SameSite`, which control the cookie's behavior and scope.

- **Host Header**: The `Host` header is mandatory in HTTP/1.1 requests. It specifies the domain name of the server (and optionally the port number) to which the request is being sent. This header is crucial for virtual hosting where multiple domains are hosted on the same IP address, and the server needs to know which domain the client is trying to access.

Here's how they might appear in an HTTP request:

```
GET /index.html HTTP/1.1
Host: www.example.com
Cookie: sessionId=abc123; username=JohnDoe
```

In this request:
- The `Host` header tells the server that the client wants to reach `www.example.com`.
- The `Cookie` header sends the `sessionId` and `username` cookies back to the server, which might use them to identify the session and customize the response based on the user's previous interactions or login state.

## Cross site request example

A cross-site request example typically involves two websites: a victim's website where the user is authenticated, and an attacker's website that tries to make unauthorized requests to the victim's website using the user's credentials. Here's a simplified example to illustrate a Cross-Site Request Forgery (CSRF) attack:

### Scenario: Transferring Money on a Banking Website

**Victim's Banking Website:**

- The user logs into their banking website `bank.com` to manage their finances.
- The website uses a form to transfer money, which includes fields for the recipient's account and the amount to transfer.
- The form submission is a POST request to an endpoint like `https://bank.com/transfer`.

**Attacker's Website:**

- The attacker creates a malicious website `evil.com` that includes an HTML form identical to the transfer money form on `bank.com`. The form on `evil.com` is set to submit to `https://bank.com/transfer`.
- The attacker tricks the user (who is logged into `bank.com` on another tab) into visiting `evil.com` and clicking a button that submits the form.
- Because the user is authenticated on `bank.com`, the browser includes the user's authentication cookies with the form submission from `evil.com`, making the request appear legitimate to `bank.com`.

**HTML Form on Attacker's Website (`evil.com`):**

```html
<form action="https://bank.com/transfer" method="POST">
  <input type="hidden" name="recipientAccount" value="attackerAccountNumber" />
  <input type="hidden" name="amount" value="1000" />
  <input type="submit" value="Click me for a surprise!" />
</form>
```

When the user clicks the button, `bank.com` receives a request to transfer money as if the user intentionally made the request, potentially resulting in unauthorized transactions.

### Mitigation:

To mitigate CSRF attacks, websites implement various strategies, such as:

- **CSRF Tokens:** Unique tokens included in forms that must be submitted with the request, making it difficult for an attacker to forge a valid request.
- **SameSite Cookie Attribute:** Restricts how cookies are sent with cross-site requests, helping to prevent CSRF attacks.
- **Checking the Referer Header:** Ensures requests are coming from allowed origins.

This example demonstrates how a cross-site request can be exploited in a CSRF attack and highlights the importance of implementing security measures to protect against such vulnerabilities.

## CORS

Cross-Origin Resource Sharing (CORS) is a security feature implemented in web browsers to prevent malicious websites from making requests to another domain without permission. It allows servers to specify who can access their resources and under what conditions. Without CORS, web browsers enforce a Same-Origin Policy (SOP) that prevents a webpage from making requests to a different domain than the one that served the webpage. This policy is in place to protect users from various types of web-based attacks, such as Cross-Site Request Forgery (CSRF).

CORS works by adding new HTTP headers that allow servers to describe the set of origins that are permitted to read that information using a web browser. Additionally, for HTTP requests that could cause side-effects on user data (such as HTTP POST, PUT, DELETE, etc.), the browser sends a "preflight" request to the server hosting the cross-origin resource, in order to check that the server will permit the actual request. In that preflight, the browser sends headers that indicate the HTTP method and headers that will be used in the actual request.

An example of a CORS header is `Access-Control-Allow-Origin`, which can be used to specify which domain is allowed to access the resource. For example, if `https://example.com` needs to request resources from `https://api.example.com`, the server at `api.example.com` can include the header `Access-Control-Allow-Origin: https://example.com` in its responses to tell the browser that the content is safe to access.

CORS is a crucial part of modern web security and is something web developers must understand and implement correctly to ensure their applications are secure and function as intended across different domains.