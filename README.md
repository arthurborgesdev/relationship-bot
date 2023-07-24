# relationship-bot

## lean engine composition

The relationship bot is made of a set o small parts that fulfill (or aim to fulfill) 80% of use cases with 20% of effort (Pareto rule). For that, it is assembled a lean engine, which is the heart of the system. It is composable of:

- User Input (User messages sent to the bot)
- Lean Engine (The system which we are going to develop)
- Bot Output (Messages sent to the user by the bot)
- Feedback mechanism of learning (inputs and outputs are fed into the system to be saved in a vector DB, which the bot remembers over time)

This is the lean engine. The smallest set of possible parts. The final system can have integration with other systems, posess more complex functionalities but by maintaining a small core, we can achieve a great level of understanding of the problem and componentization. 

### parts

The innital parts for the lean engine are:

- Main language: Go
- AI LLM: go-openai package
- Vector DB: milvus-sdk-go package (we will be using milvus for start, but later this layer can be abstracted to connect to any vector db or other storage technology)
- IO: fiber package (the desired implementation is for whatsapp chatbots. But to remove the initial burden and make the engine as extensible as possible, we will be using HTTP requests to communicate to and from the system. In this scenario, the communication layer is stateless, but the lean engine is stateful)
