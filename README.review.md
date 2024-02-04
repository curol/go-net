# Review

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