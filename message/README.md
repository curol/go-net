
# Message

## Stream
In computer science, a stream is a sequence of data elements made available over time. **A stream can be thought of as a continuous flow of data that can be read or written in small chunks, rather than as a fixed-size block of data.**

In Go, a stream is often represented using an io.Reader or io.Writer interface. An io.Reader represents a source of data that can be read in small chunks, while an io.Writer represents a destination for data that can be written in small chunks.

For example, in the code you provided, buffer is a bytes.Buffer that represents a stream of bytes that can be read or written using an io.Reader or io.Writer. The bufio.Reader and bufio.Writer types are used to read and write data to the buffer stream in small chunks.

Streams are often used in networking and file I/O, where data is read or written in small chunks over a potentially slow or unreliable connection. By using streams, you can avoid loading large amounts of data into memory at once, and can handle data as it becomes available.

## Head

## Body
		 Body is the payload of the message.
		
		 For small data, the entire body can be read at once.
		 For large data, the body can be streamed in chunks.
		 For example, a file can be streamed in chunks.
		 For example, a video can be streamed in chunks.
		 For example, a large JSON object can be streamed in chunks.
		
		 Why carefully read the body with content-length?
		 Because the size of the data might be too big for server to read all at once.
		 For streams, we don't want to read the entire body at once into memory.
		 If its too big, we might run out of memory.
		 Instead, implement strategy for different sizes of the data and use chunks or buffers to read from the connection.


## Reader

The io.Reader interface in Go is a fundamental interface that represents the read end of a stream of data.
It has a single method, Read, which reads up to len(p) bytes into p.

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

The Read method takes a byte slice as an argument, fills it with data and returns the number of bytes read and an error, if any occurred.
If there is no more data to be read, Read returns io.EOF (end-of-file) error.

Here's how you might use it:

```go
buf := make([]byte, 1024)
n, err := reader.Read(buf)
if err != nil && err != io.EOF {
    panic(err)
}
fmt.Println("Bytes read:", n)
```

In this example, reader is an io.Reader. We create a buffer buf and pass it to reader.Read.
The Read method fills buf with data and returns the number of bytes read.
If an error occurs during the read operation and it's not io.EOF, we panic.
Otherwise, we print the number of bytes read.

### ReadCloser
ReadCloser is the interface that groups the basic Read and Close methods.
	
```
	 type ReadCloser interface {
	   Reader
	 	 Closer
	 }
```


## Writer

In Go, the io.Writer interface is a fundamental interface that represents the write end of a stream of data.
It has a single method, Write, which writes some data and returns the number of bytes written and an error, if any occurred.

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

The Write method takes a byte slice as input, writes as much of it as possible, and returns the number of bytes it wrote and an error value.
If the number of bytes written is less than the length of the input byte slice, it should return an error explaining why it couldn't write the whole slice.

The Flush method is not part of the io.Writer interface, but it is often found in buffered writers like bufio.Writer. The purpose of Flush is to clear the buffer, writing any buffered data to the underlying writer.

```go
type Writer interface {
    Write(p []byte) (n int, err error)
    Flush() error
}
```

When you write data to a buffered writer, the data is not usually written immediately to the underlying writer.
Instead, it's held in a buffer, and only written when the buffer becomes full.
This can improve performance by reducing the number of write operations.
However, sometimes you need to ensure that all written data is sent immediately, without waiting for the buffer to fill up.
That's when you would call Flush.

## WriterTo
GitHub Copilot: In Go, the `io.WriterTo` interface is used to write data to a destination. It is similar to the `io.Writer` interface, but instead of writing data to a stream, it writes data to another `io.Writer`.

Here's an example of how to use the `io.WriterTo` interface to write data to a file:

```go
package main

import (
	"io"
	"os"
)

type MyData struct {
	Data string
}

func (d MyData) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte(d.Data))
	return int64(n), err
}

func main() {
	// Create a new instance of MyData
	data := MyData{Data: "Hello, world!"}

	// Open a file for writing
	file, err := os.Create("output.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write the data to the file using WriteTo
	_, err = data.WriteTo(file)
	if err != nil {
		panic(err)
	}
}
```

In this example, we define a custom type `MyData` that has a `WriteTo` method that implements the `io.WriterTo` interface. The `WriteTo` method takes an `io.Writer` as an argument and writes the data to the writer using `w.Write`.

We then create a new instance of `MyData` and open a file for writing using `os.Create`. We write the data to the file using the `WriteTo` method of `MyData`.

Note that in this example, we are writing the entire data to the file at once. In practice, you may want to write the data in smaller chunks to avoid consuming too much memory or overwhelming the destination writer.

## Parsing vs Serialization

The opposite of parsing is usually referred to as "serialization". Parsing is the process of analyzing a string or data structure to extract meaningful information, while serialization is the process of converting an object or data structure into a format that can be stored or transmitted.

In other words, parsing is about extracting data from a structured format, while serialization is about converting data into a structured format.

## Functions 

### Pure

GitHub Copilot: A pure function is a function that always produces the same output given the same input, and has no side effects. In other words, a pure function does not modify any state outside of its own scope, and does not rely on any state outside of its own scope.

Pure functions are deterministic, meaning that they always produce the same output given the same input. This makes them easy to reason about and test, because you can be sure that the output of a pure function will not change unexpectedly.

Pure functions are also composable, meaning that you can combine them to create more complex functions. Because pure functions have no side effects, you can safely chain them together without worrying about unexpected interactions between them.

Here's an example of a pure function in Go:

```go
package main

import "fmt"

func add(a, b int) int {
    return a + b
}

func main() {
    x := add(2, 3) // x is 5
    y := add(2, 3) // y is also 5
    fmt.Println(x, y)
}
```

In this example, the `add` function takes two integers as input and returns their sum. The `add` function has no side effects and always produces the same output given the same input, so it is a pure function.

We call the `add` function twice with the same input, and get the same output both times. This demonstrates the deterministic nature of pure functions.

### Impure

GitHub Copilot: A function with side-effects is called an impure function. An impure function is a function that modifies state outside of its own scope, or relies on state outside of its own scope. Impure functions can have unpredictable behavior and can make code harder to reason about and test.

Types of impure functions:

1. Functions with side effects: These are functions that modify state outside of their own scope, such as global variables or files. Examples include functions that write to a file or modify a global variable.

2. Functions that rely on external state: These are functions that rely on state outside of their own scope, such as the current time or the state of a network connection. Examples include functions that generate random numbers or make network requests.

3. Functions that are non-deterministic: These are functions that do not always produce the same output given the same input. Examples include functions that generate random numbers or use the current time as part of their output.

4. Functions that throw exceptions: These are functions that can throw exceptions or errors, which can cause unexpected behavior in the calling code. Examples include functions that divide by zero or access an out-of-bounds array index.

Note that impure functions can be useful in some cases, such as when you need to modify state outside of the function's scope. However, impure functions can also make code harder to reason about and test, because their behavior can be unpredictable.

### Mutators

GitHub Copilot: Mutator methods are methods that modify the state of an object or struct. In Go, mutator methods are typically defined as methods with a pointer receiver that modify the fields of the struct directly.

Mutator methods are often used in conjunction with accessor methods (also known as getters and setters) to provide controlled access to the fields of a struct. By defining mutator methods that modify the fields of a struct directly, you can ensure that the struct's invariants are maintained.

Here's an example of a mutator method in Go:

```go
package main

import "fmt"

type Person struct {
    name string
}

func (p *Person) SetName(name string) {
    p.name = name
}

func main() {
    p := &Person{name: "John"}
    fmt.Println(p.name) // Output: John
    p.SetName("Jane")
    fmt.Println(p.name) // Output: Jane
}
```

In this example, we define a `Person` struct with a private `name` field. We define a `SetName` method that takes a `name` parameter and sets the value of the `name` field. We then create a `Person` instance and use the `SetName` method to set the `name` field.

Note that the `SetName` method modifies the `name` field directly, so it is a mutator method.

### Getters and Setters
GitHub Copilot: Getters and setters are methods or functions that are used to access or modify the values of private fields in an object or struct. 

A getter is a method that returns the value of a private field. It is used to provide read-only access to the field. Getters are typically named using the format `Get<Field>()`, where `<Field>` is the name of the field with the first letter capitalized.

A setter is a method that sets the value of a private field. It is used to provide write-only access to the field. Setters are typically named using the format `Set<Field>()`, where `<Field>` is the name of the field with the first letter capitalized.

In Go, getters and setters are not typically used for simple struct fields, because Go does not have a built-in way to enforce encapsulation. Instead, it is common to use exported fields (fields with the first letter capitalized) and provide methods that operate on the struct as a whole.

For example, instead of using getters and setters for a struct with a private field `name`, you might define a method `GetName()` that returns the value of the `name` field, and a method `SetName(name string)` that sets the value of the `name` field.

Here's an example of how to define getters and setters for a struct in Go:

```go
package main

import "fmt"

type Person struct {
    name string
}

func (p *Person) GetName() string {
    return p.name
}

func (p *Person) SetName(name string) {
    p.name = name
}

func main() {
    p := &Person{name: "John"}
    fmt.Println(p.GetName()) // Output: John
    p.SetName("Jane")
    fmt.Println(p.GetName()) // Output: Jane
}
```

In this example, we define a `Person` struct with a private `name` field. We define a `GetName` method that returns the value of the `name` field, and a `SetName` method that sets the value of the `name` field. We then create a `Person` instance and use the `GetName` and `SetName` methods to get and set the `name` field.

## What is parsing?

Parsing is the process of analyzing a string of data (such as a text file or network message) to determine its structure and extract meaningful information from it. In the context of programming, parsing is often used to convert data from one format to another, or to extract specific pieces of information from a larger data set.

For example, in the excerpt from util.go that you provided, the parseRequestLine function parses the first line of an HTTP request to extract the method, URL, and HTTP version. The function reads the first line from a bufio.Reader, analyzes its structure, and returns a parsedRequestLine struct with the relevant information.

Parsing can be a complex process, especially when dealing with data that has a complex or ambiguous structure. Many programming languages provide built-in parsing libraries or frameworks to simplify the process of parsing data.

##  What is a delimeter?

GitHub Copilot: A delimiter (or separator) is a character or sequence of characters that is used to separate fields or values in a data stream or file. In the context of the `ToString` method in the `HashMap` struct in the `hashmap.go` excerpt you provided, the `delm` parameter is a delimiter that is used to separate the key-value pairs in the resulting string representation of the `HashMap`.

For example, if the `delm` parameter is set to `,`, the resulting string might look like this:

```
key1:value1,key2:value2,key3:value3,
```

In this example, the `:` character is used as a separator between the keys and values, and the `,` character is used as a separator between the key-value pairs.