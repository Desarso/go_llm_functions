I really like the syntax of ell, especially with decorators. Here's an example:

    @ell.simple(model="gpt-4")
    def hello(name: str):
        """You are a helpful assistant."""
        return f"Say hello to {name}!"

    
However, Go doesn't have decorators, so I came up with something like this:

    func sayHi(user string) string {
        return fmt.Sprintf("Say hello to the user of name %s", user)
    }
    
    // You call the function like this:
    llm.Ell(sayHi)("Steven")
    
This approach still lets you define LLM functions as simple functions that take a string and return a string.

I’d like to expand this further by adding features such as system messages, tool calls, streaming, and more. One challenge is figuring out the best syntax for these additions.

Unlike decorators, this design doesn’t enforce the function as an LLM function—you can’t call it directly. It adds some flexibility but also requires careful consideration for usability and consistency.

After workshopping it with AI, I think a good sytanx is something like this: 

    // Example: Only use the handler, no message or options
    sayHi := Register(
        func(systemMessage, user string) string {
            return fmt.Sprintf("%s Say hello to %s!", systemMessage, user)
        },
    )
    
    // Example: Provide a system message only
    sayHi := Register(
        func(systemMessage, user string) string {
            return fmt.Sprintf("%s Say hello to %s!", systemMessage, user)
        },
        "You are a friendly assistant.", // Only system message
    )
    
    // Example: Provide options only
    sayHi := Register(
        func(systemMessage, user string) string {
            return fmt.Sprintf("%s Say hello to %s!", systemMessage, user)
        },
        Options{Top: 3}, // Only options
    )
    
    // Example: Provide both system message and options
    sayHi := Register(
        func(systemMessage, user string) string {
            return fmt.Sprintf("%s Say hello to %s!", systemMessage, user)
        },
        "You are a friendly assistant.", // SystemMessage
        Options{                         // Options object
            Debug: true,
            Top:   5,
        },
    )
So I wrap the function definition with Register, maybe call it something else. So later I just call it.
I can rename Register to llm. Get rid of the systemMessage on every function. This perfect I think. I can select the model in options.
I can also change the api key, default model, etc, from the module variables. 
