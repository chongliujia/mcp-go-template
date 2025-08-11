#!/usr/bin/env python3
"""
LangGraph Agent to test MCP service integration
"""

import asyncio
import json
import uuid
from typing import Dict, Any, List, Optional
from dataclasses import dataclass
from enum import Enum

import httpx
import websockets
from langgraph.graph import StateGraph, START, END
from langgraph.graph.state import CompiledStateGraph
from typing_extensions import TypedDict


class MCPServiceStatus(Enum):
    UNKNOWN = "unknown"
    AVAILABLE = "available"
    UNAVAILABLE = "unavailable"


@dataclass
class MCPTool:
    name: str
    description: str
    input_schema: Dict[str, Any]


class AgentState(TypedDict):
    messages: List[Dict[str, Any]]
    mcp_status: MCPServiceStatus
    available_tools: List[MCPTool]
    current_task: Optional[str]
    result: Optional[Dict[str, Any]]
    error: Optional[str]


class MCPTestAgent:
    def __init__(self, mcp_server_url: str = "localhost:8030"):
        self.mcp_server_url = mcp_server_url
        self.http_client = httpx.AsyncClient(timeout=30.0)
        self.websocket = None
        self.graph = self._build_graph()

    async def _send_mcp_request(self, method: str, params: Dict[str, Any] = None) -> Dict[str, Any]:
        """Send JSON-RPC request via WebSocket"""
        if not self.websocket:
            ws_url = f"ws://{self.mcp_server_url}/mcp"
            self.websocket = await websockets.connect(ws_url)
        
        request = {
            "jsonrpc": "2.0",
            "id": str(uuid.uuid4()),
            "method": method,
            "params": params or {}
        }
        
        await self.websocket.send(json.dumps(request))
        response = await self.websocket.recv()
        return json.loads(response)

    def _build_graph(self) -> CompiledStateGraph:
        """Build the LangGraph workflow"""
        workflow = StateGraph(AgentState)
        
        # Add nodes
        workflow.add_node("check_mcp_status", self._check_mcp_status)
        workflow.add_node("discover_tools", self._discover_tools)
        workflow.add_node("test_calculator", self._test_calculator)
        workflow.add_node("test_web_search", self._test_web_search)
        workflow.add_node("test_knowledge_graph", self._test_knowledge_graph)
        workflow.add_node("generate_report", self._generate_report)
        
        # Add edges
        workflow.add_edge(START, "check_mcp_status")
        
        workflow.add_conditional_edges(
            "check_mcp_status",
            self._route_after_status_check,
            {
                "discover_tools": "discover_tools",
                "end": END
            }
        )
        
        workflow.add_edge("discover_tools", "test_calculator")
        workflow.add_edge("test_calculator", "test_web_search")
        workflow.add_edge("test_web_search", "test_knowledge_graph")
        workflow.add_edge("test_knowledge_graph", "generate_report")
        workflow.add_edge("generate_report", END)
        
        return workflow.compile()

    async def _check_mcp_status(self, state: AgentState) -> AgentState:
        """Check if MCP service is available"""
        try:
            # First check HTTP health endpoint
            response = await self.http_client.get(f"http://{self.mcp_server_url}/health")
            if response.status_code == 200:
                # Then test WebSocket connection with initialize
                init_response = await self._send_mcp_request("initialize", {
                    "protocolVersion": "2024-11-05",
                    "capabilities": {},
                    "clientInfo": {
                        "name": "langgraph-test-agent",
                        "version": "1.0.0"
                    }
                })
                
                if init_response.get("result"):
                    # Send initialized notification to complete handshake
                    await self.websocket.send(json.dumps({
                        "jsonrpc": "2.0",
                        "method": "initialized"
                    }))
                    
                    state["mcp_status"] = MCPServiceStatus.AVAILABLE
                    state["messages"].append({
                        "type": "status",
                        "content": "MCP service is available ✓"
                    })
                else:
                    state["mcp_status"] = MCPServiceStatus.UNAVAILABLE
                    state["error"] = f"MCP initialization failed: {init_response.get('error')}"
            else:
                state["mcp_status"] = MCPServiceStatus.UNAVAILABLE
                state["error"] = f"MCP service returned status {response.status_code}"
        except Exception as e:
            state["mcp_status"] = MCPServiceStatus.UNAVAILABLE
            state["error"] = f"Failed to connect to MCP service: {str(e)}"
            state["messages"].append({
                "type": "error",
                "content": f"MCP service connection failed: {str(e)}"
            })
        
        return state

    def _route_after_status_check(self, state: AgentState) -> str:
        """Route based on MCP service status"""
        if state["mcp_status"] == MCPServiceStatus.AVAILABLE:
            return "discover_tools"
        else:
            return "end"

    async def _discover_tools(self, state: AgentState) -> AgentState:
        """Discover available MCP tools"""
        try:
            response = await self._send_mcp_request("tools/list")
            
            if response.get("result"):
                tools_data = response["result"]
                tools = []
                for tool in tools_data.get("tools", []):
                    tools.append(MCPTool(
                        name=tool["name"],
                        description=tool["description"],
                        input_schema=tool["inputSchema"]
                    ))
                
                state["available_tools"] = tools
                state["messages"].append({
                    "type": "info",
                    "content": f"Discovered {len(tools)} tools: {[t.name for t in tools]}"
                })
            else:
                state["error"] = f"Failed to discover tools: {response.get('error')}"
        except Exception as e:
            state["error"] = f"Tool discovery failed: {str(e)}"
        
        return state

    async def _test_calculator(self, state: AgentState) -> AgentState:
        """Test calculator tool"""
        state["current_task"] = "Testing calculator tool"
        
        try:
            response = await self._send_mcp_request("tools/call", {
                "name": "calculator",
                "arguments": {
                    "expression": "2 + 3 * 4"
                }
            })
            
            if response.get("result"):
                state["messages"].append({
                    "type": "test_result",
                    "tool": "calculator",
                    "input": "2 + 3 * 4",
                    "output": response["result"],
                    "status": "success"
                })
            else:
                state["messages"].append({
                    "type": "test_result",
                    "tool": "calculator",
                    "status": "failed",
                    "error": str(response.get("error"))
                })
        except Exception as e:
            state["messages"].append({
                "type": "test_result",
                "tool": "calculator",
                "status": "error",
                "error": str(e)
            })
        
        return state

    async def _test_web_search(self, state: AgentState) -> AgentState:
        """Test web search tool"""
        state["current_task"] = "Testing web search tool"
        
        try:
            response = await self._send_mcp_request("tools/call", {
                "name": "web_search",
                "arguments": {
                    "query": "LangGraph tutorial",
                    "max_results": 3
                }
            })
            
            if response.get("result"):
                state["messages"].append({
                    "type": "test_result",
                    "tool": "web_search",
                    "input": "LangGraph tutorial",
                    "output": response["result"],
                    "status": "success"
                })
            else:
                state["messages"].append({
                    "type": "test_result",
                    "tool": "web_search",
                    "status": "failed",
                    "error": str(response.get("error"))
                })
        except Exception as e:
            state["messages"].append({
                "type": "test_result",
                "tool": "web_search",
                "status": "error",
                "error": str(e)
            })
        
        return state

    async def _test_knowledge_graph(self, state: AgentState) -> AgentState:
        """Test knowledge graph tool"""
        state["current_task"] = "Testing knowledge graph tool"
        
        try:
            response = await self._send_mcp_request("tools/call", {
                "name": "knowledge_graph",
                "arguments": {
                    "query": "What is artificial intelligence?",
                    "operation": "search"
                }
            })
            
            if response.get("result"):
                state["messages"].append({
                    "type": "test_result",
                    "tool": "knowledge_graph",
                    "input": "What is artificial intelligence?",
                    "output": response["result"],
                    "status": "success"
                })
            else:
                state["messages"].append({
                    "type": "test_result",
                    "tool": "knowledge_graph",
                    "status": "failed",
                    "error": str(response.get("error"))
                })
        except Exception as e:
            state["messages"].append({
                "type": "test_result",
                "tool": "knowledge_graph",
                "status": "error",
                "error": str(e)
            })
        
        return state

    async def _generate_report(self, state: AgentState) -> AgentState:
        """Generate test report"""
        state["current_task"] = "Generating test report"
        
        test_results = [msg for msg in state["messages"] if msg.get("type") == "test_result"]
        
        report = {
            "mcp_service_status": state["mcp_status"].value,
            "total_tools_discovered": len(state.get("available_tools", [])),
            "tests_conducted": len(test_results),
            "successful_tests": len([r for r in test_results if r.get("status") == "success"]),
            "failed_tests": len([r for r in test_results if r.get("status") in ["failed", "error"]]),
            "detailed_results": test_results
        }
        
        state["result"] = report
        state["messages"].append({
            "type": "report",
            "content": "Test report generated successfully"
        })
        
        return state

    async def run_tests(self) -> Dict[str, Any]:
        """Run all MCP service tests"""
        initial_state = AgentState(
            messages=[{"type": "info", "content": "Starting MCP service tests..."}],
            mcp_status=MCPServiceStatus.UNKNOWN,
            available_tools=[],
            current_task=None,
            result=None,
            error=None
        )
        
        final_state = await self.graph.ainvoke(initial_state)
        return final_state

    async def __aenter__(self):
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.websocket:
            await self.websocket.close()
        await self.http_client.aclose()


async def main():
    """Main execution function"""
    print("🚀 Starting LangGraph MCP Agent Test...")
    
    async with MCPTestAgent() as agent:
        final_state = await agent.run_tests()
        
        print("\n" + "="*60)
        print("📊 MCP SERVICE TEST REPORT")
        print("="*60)
        
        if final_state.get("result"):
            report = final_state["result"]
            print(f"MCP Service Status: {report['mcp_service_status']}")
            print(f"Tools Discovered: {report['total_tools_discovered']}")
            print(f"Tests Conducted: {report['tests_conducted']}")
            print(f"Successful Tests: {report['successful_tests']}")
            print(f"Failed Tests: {report['failed_tests']}")
            
            print("\n📝 Detailed Results:")
            for result in report["detailed_results"]:
                status_emoji = "✅" if result["status"] == "success" else "❌"
                print(f"{status_emoji} {result['tool']}: {result['status']}")
                if result.get("error"):
                    print(f"   Error: {result['error']}")
        
        if final_state.get("error"):
            print(f"\n❌ Error: {final_state['error']}")
        
        print("\n💬 Execution Log:")
        for msg in final_state.get("messages", []):
            if msg["type"] == "info":
                print(f"ℹ️  {msg['content']}")
            elif msg["type"] == "error":
                print(f"❌ {msg['content']}")
            elif msg["type"] == "status":
                print(f"🔄 {msg['content']}")


if __name__ == "__main__":
    asyncio.run(main())