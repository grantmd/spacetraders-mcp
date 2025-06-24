# AcceptContract Tool

The `accept_contract` tool allows you to accept contracts in SpaceTraders. When you accept a contract, you commit to fulfilling its terms and receive an upfront payment.

## Usage

### Tool Definition

- **Name**: `accept_contract`
- **Description**: Accept a contract by its ID. This commits the agent to fulfilling the contract terms and provides an upfront payment.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `contract_id` | string | Yes | The unique identifier of the contract to accept |

### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "accept_contract",
    "arguments": {
      "contract_id": "clm0n4k8q0001js08g2h1k9v8"
    }
  }
}
```

### Example Response (Success)

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\n  \"success\": true,\n  \"message\": \"Successfully accepted contract clm0n4k8q0001js08g2h1k9v8\",\n  \"contract\": {\n    \"id\": \"clm0n4k8q0001js08g2h1k9v8\",\n    \"faction\": \"COSMIC\",\n    \"type\": \"PROCUREMENT\",\n    \"accepted\": true,\n    \"fulfilled\": false,\n    \"expiration\": \"2024-12-31T23:59:59Z\",\n    \"terms\": {\n      \"deadline\": \"2024-12-30T23:59:59Z\",\n      \"payment\": {\n        \"on_accepted\": 10000,\n        \"on_fulfilled\": 50000\n      },\n      \"deliver\": [\n        {\n          \"tradeSymbol\": \"IRON_ORE\",\n          \"destinationSymbol\": \"X1-COSMIC-STATION\",\n          \"unitsRequired\": 100,\n          \"unitsFulfilled\": 0\n        }\n      ]\n    }\n  },\n  \"agent\": {\n    \"symbol\": \"MYAGENT\",\n    \"credits\": 110000,\n    \"ships\": 1,\n    \"faction\": \"COSMIC\"\n  }\n}"
      }
    ],
    "isError": false
  }
}
```

### Example Response (Error)

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Failed to accept contract: API request failed with status 404: Contract not found"
      }
    ],
    "isError": true
  }
}
```

## Response Fields

When successful, the tool returns a JSON object with the following structure:

### Contract Object

- `id`: The contract's unique identifier
- `faction`: The faction that issued the contract
- `type`: The type of contract (e.g., "PROCUREMENT")
- `accepted`: Whether the contract has been accepted (always true after acceptance)
- `fulfilled`: Whether the contract has been fulfilled
- `expiration`: When the contract expires
- `terms`: Contract terms including deadline, payment, and delivery requirements

### Agent Object

- `symbol`: Your agent's symbol/name
- `credits`: Your agent's current credits (updated after receiving the acceptance payment)
- `ships`: Number of ships in your fleet
- `faction`: Your agent's starting faction

## Prerequisites

1. You must have a valid SpaceTraders API token configured
2. The contract must exist and be available for acceptance
3. The contract must not already be accepted
4. The contract must not be expired

## Common Error Cases

- **Contract not found**: The provided contract ID doesn't exist
- **Contract already accepted**: You've already accepted this contract
- **Contract expired**: The deadline to accept has passed
- **Invalid contract ID**: The contract ID format is invalid
- **Missing or empty contract ID**: No contract ID was provided

## Related Resources

- Use the [contracts resource](../resources/contracts.md) to list available contracts
- Check contract details before accepting to understand the requirements
- Monitor contract deadlines to avoid expiration

## Notes

- Accepting a contract provides an immediate payment (on_accepted amount)
- You'll receive additional payment when the contract is fulfilled (on_fulfilled amount)
- Once accepted, you're committed to fulfilling the contract terms
- Failed contracts may impact your reputation with the issuing faction