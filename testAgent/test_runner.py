#!/usr/bin/env python3
"""
Simple test runner for MCP service
"""

import asyncio
import sys
import os
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from langgraph_mcp_agent import MCPTestAgent


async def quick_test():
    """Quick test of MCP service availability"""
    print("🔍 Quick MCP Service Test")
    print("-" * 30)
    
    async with MCPTestAgent() as agent:
        # Just test connectivity
        from langgraph_mcp_agent import AgentState, MCPServiceStatus
        
        state = AgentState(
            messages=[],
            mcp_status=MCPServiceStatus.UNKNOWN,
            available_tools=[],
            current_task=None,
            result=None,
            error=None
        )
        
        # Check service status
        result_state = await agent._check_mcp_status(state)
        
        if result_state["mcp_status"] == MCPServiceStatus.AVAILABLE:
            print("✅ MCP service is running and accessible")
            
            # Try to discover tools
            tools_state = await agent._discover_tools(result_state)
            if tools_state.get("available_tools"):
                print(f"✅ Found {len(tools_state['available_tools'])} tools:")
                for tool in tools_state["available_tools"]:
                    print(f"   • {tool.name}: {tool.description}")
            else:
                print("⚠️  No tools discovered")
        else:
            print("❌ MCP service is not available")
            if result_state.get("error"):
                print(f"   Error: {result_state['error']}")


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "quick":
        asyncio.run(quick_test())
    else:
        # Import and run full test
        from langgraph_mcp_agent import main
        asyncio.run(main())