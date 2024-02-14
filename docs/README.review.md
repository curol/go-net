# Review

## Vim

### Paste
    In Vim, you can paste the contents of the clipboard or a previously yanked (copied) text into the next line using the `p` command in normal mode. Here's how:

1. First, copy some text. You can do this in several ways:
   - Use `yy` to yank (copy) the current line.
   - Use `y` followed by a movement command to yank some other text. For example, `yw` yanks the current word.

2. Move the cursor to the line after which you want to paste the text.

3. Press `p` to paste the text after the cursor (on the next line if you've yanked a whole line).

If you want to paste before the current line, you can use `P` instead of `p`.

Remember, Vim has its own clipboard for yanking and pasting text. If you want to paste text from the system clipboard, you might need to use a different command, depending on your Vim configuration and the system you're using. For example, in many configurations, you can use `"*p` to paste from the system clipboard.
## Listners

The number of listeners you need for a server depends on the architecture of your application and the requirements of your system.

In a typical web server setup, you would only need one listener. This listener would accept incoming connections and then hand them off to a worker (which could be a thread, a goroutine, etc.) to handle the request. This allows the listener to go back to accepting more connections as soon as possible.

However, there are scenarios where you might want more than one listener. For example, if you're building a server that needs to listen on multiple ports (for example, one for HTTP and one for HTTPS), you would need a separate listener for each port.

In a more complex system, you might have multiple listeners for different services, each running in its own process or even on its own machine. This is common in microservices architectures, where each service is a separate component that communicates over the network.

In general, the number of listeners is less about the number of clients you expect to connect and more about the architecture of your server and the protocols you're using.

## Wire Format

In the context of programming and data communication, a "wire format" refers to the way data is serialized or structured for transmission over a network, also known as "on-the-wire". This term is used because the data is often sent over physical wires, like Ethernet cables, although it can also be sent over wireless connections.

The wire format is important because it ensures that both the sender and receiver of the data understand how to interpret it. It defines things like:

- The order in which bytes are sent and received (endianness)
- How complex data types are represented
- How data is delimited or separated
- How data is encoded (for example, ASCII, UTF-8, binary)

For example, the JSON and XML data formats are both wire formats. They define how to represent complex data types like arrays and objects (in JSON) or elements and attributes (in XML) as text.

Protobuf (Protocol Buffers) by Google is another example of a wire format, but unlike JSON or XML which are text-based, Protobuf is binary, which makes it more compact and faster to encode/decode.

It's important to note that a good wire format is typically compact (to reduce transmission time and storage requirements), fast to serialize and deserialize, and platform-independent.