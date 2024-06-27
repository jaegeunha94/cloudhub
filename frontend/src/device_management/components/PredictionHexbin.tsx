import React, {useEffect, useMemo, useRef, useState} from 'react'
import * as d3 from 'd3'
import {hexbin} from 'd3-hexbin'
import FancyScrollbar from 'src/shared/components/FancyScrollbar'
import {DEFAULT_CELL_BG_COLOR} from 'src/dashboards/constants'

interface HexagonData {
  x: number
  y: number
  statusColor: string
  hostname: string
}

interface HexagonInputData {
  statusColor: string
  hostname: string
}

interface GenerateHexagonData {
  x: number
  y: number
}

//핵사곤 크기
const hexRadius = 30
const hexPadding = 5

const PredictionHexbin: React.FC = () => {
  const svgRef = useRef<SVGSVGElement>(null)

  const [colorChange, setColorChange] = useState<number>(2)

  //   for (let i = 0; i < 430; i++) {
  //     const newHostname = `${i + 1}`
  //     inputData.push({
  //       statusColor: i % colorChange === 1 ? 'red' : 'green',
  //       hostname: newHostname,
  //     })
  //   }

  const inputData = useMemo<HexagonInputData[]>(() => {
    return [...Array(150)].map((_, i) => {
      const newHostname = `${i + 1}`
      return {
        statusColor: i % colorChange === 1 ? 'red' : 'green',
        hostname: newHostname,
      }
    })
  }, [colorChange])

  const attachEventHandlers = () => {
    if (!svgRef.current) return
    const svg = d3.select(svgRef.current)
    svg
      .selectAll('.hexagon')
      .on('mouseover', function () {
        d3.select(this)
          .transition()
          .duration(150)
          .attr('transform', d => `translate(${d.x},${d.y}) scale(1.1)`)
          .style('cursor', 'pointer')
      })
      .on('mouseout', function () {
        d3.select(this)
          .transition()
          .duration(150)
          .attr('transform', d => `translate(${d.x},${d.y})`)
          .style('cursor', 'default')
      })
  }

  const generateHexagonData = (): (GenerateHexagonData &
    HexagonInputData)[] => {
    const hexagonData = []
    const svgWidth = svgRef.current.clientWidth
    const hexWidth = Math.sqrt(3) * hexRadius // Hexagon의 폭
    const hexHeight = 2 * hexRadius // Hexagon의 높이
    let xOffset = 0
    let yOffset = 0
    for (let i = 0; i < inputData.length; i++) {
      const isOddRow = Math.floor(yOffset / ((hexHeight * 3) / 4)) % 2 === 1
      const x = xOffset + (isOddRow ? hexRadius * 1.5 : 0)
      const y = yOffset + hexHeight * 0.75
      hexagonData.push({
        x: x + hexPadding,
        y: y + hexPadding,
        ...inputData[i],
      })
      if (x + hexWidth > svgWidth - hexHeight) {
        yOffset += (hexHeight * 3) / 4
        xOffset = 0
      } else {
        xOffset += hexWidth
      }
    }

    return hexagonData
  }

  useEffect(() => {
    const resizeObserver = new ResizeObserver(_ => {
      generateHexagonData()
      drawHexagons()
      attachEventHandlers()
    })
    if (svgRef.current) {
      resizeObserver.observe(svgRef.current.parentNode as Element)
    }
    return () => {
      resizeObserver.disconnect()
    }
  }, [])

  const drawHexagons = () => {
    if (!svgRef.current) return
    const svg = d3.select(svgRef.current)
    svg.selectAll('*').remove()

    const hexagonData = generateHexagonData()
    const svgWidth = svgRef.current.clientWidth
    const svgHeight = svgRef.current.clientHeight
    const maxHeight =
      hexagonData.reduce((max, d) => Math.max(max, d.y), 0) + hexRadius
    svgRef.current.style.height = `${maxHeight}px`
    const hexbinGenerator = hexbin<HexagonData>()
      .extent([
        [0, 0],
        [svgWidth, svgHeight],
      ])
      .radius(hexRadius)
      .x(d => d.x)
      .y(d => d.y)

    svg
      .selectAll('.hexagon')
      .data(hexbinGenerator(hexagonData))
      .enter()
      .append('path')
      .attr('class', 'hexagon')
      .attr('d', hexbinGenerator.hexagon(hexRadius - hexPadding))
      .attr('transform', d => `translate(${d.x},${d.y})`)
      .attr('fill', d => d[0]?.statusColor)

    svg
      .selectAll('.hexagon-text')
      .data(hexagonData)
      .enter()
      .append('text')
      .attr('class', 'hexagon-text')
      .attr('x', d => d.x)
      .attr('y', d => d.y)
      .attr('text-anchor', 'middle')
      .attr('dx', '.35em')
  }

  //초기화
  useEffect(() => {
    drawHexagons()
    attachEventHandlers()
  }, [colorChange])

  return (
    <FancyScrollbar style={{height: 'calc(100% - 45px)'}} autoHide={true}>
      <div
        style={{
          backgroundColor: DEFAULT_CELL_BG_COLOR,
          height: 'calc(100% - 45px)',
        }}
        className={'tab-pannel'}
      >
        <svg ref={svgRef} style={{width: '100%', height: '80%'}}></svg>
      </div>
    </FancyScrollbar>
  )
}
export default PredictionHexbin
