import dagre from '@dagrejs/dagre'
import { Edge, Node, Position } from '@xyflow/react'

const dagreGraph = new dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}))

const nodeWidth = 500 // TODO maybe update at some point
const nodeHeight = 500
const direction = 'LR'

interface getLayoutedElementsProps {
  nodes: Node[]
  edges: Edge[]
}

export const getLayoutedElements = ({
  nodes,
  edges,
}: getLayoutedElementsProps): { nodes: Node[]; edges: Edge[] } => {
  dagreGraph.setGraph({ rankdir: direction })

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight })
  })

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target)
  })

  dagre.layout(dagreGraph)

  const newNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id)
    const newNode = {
      ...node,
      targetPosition: Position.Top,
      sourcePosition: Position.Bottom,
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
    }

    return newNode
  })

  return { nodes: newNodes, edges: edges }
}
