import React, {useEffect, useMemo, useState} from 'react'

// Type
import {GetLearningDLData} from 'src/types'

// API
import {getLearningRstDL} from 'src/device_management/apis'

// Components
import {DLNxRstChart} from 'src/device_management/components/PredictionModalDLContent'

interface Props {
  host: string
}
function PredictionModalDLChart({host}: Props) {
  const [dlResultData, setDlResultData] = useState<GetLearningDLData>()

  const [noData, setNoData] = useState(true)

  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getDlData(host)
  }, [host])

  const getDlData = async (host: string) => {
    try {
      const result = await getLearningRstDL(host)
      if (!result) {
        setNoData(true)
      } else {
        setDlResultData(result.data)
        setNoData(false)
      }
    } catch (e) {
      console.error(e)
    }
    setLoading(false)
  }

  //Dl-tran,valid First Chart

  const dlTrainData = useMemo(() => {
    return {
      label: 'Train loss',
      data: dlResultData?.train_loss?.map(item => {
        return item
      }),
      backgroundColor: 'rgba(45, 199, 232, 1)',
      order: 2,
      pointRadius: 0,
      borderWidth: 1,

      borderColor: 'rgba(45, 199, 232, 1)',
    }
  }, [dlResultData?.train_loss])

  const dlValidData = useMemo(() => {
    return {
      label: 'Valid lLoss',
      data: dlResultData?.valid_loss?.map(item => {
        return item
      }),

      backgroundColor: 'rgba(255, 166, 77, 1)',
      order: 2,
      pointRadius: 0,
      borderWidth: 1,
      borderColor: 'rgba(255, 166, 77, 1)',
    }
  }, [dlResultData?.valid_loss])

  const trainChartDataSet = useMemo(() => {
    return {
      datasets: [dlTrainData, dlValidData],
      labels: dlResultData?.valid_loss.map((_, i) => i),
    }
  }, [dlTrainData, dlValidData])

  //DL-MSE Second Data
  const dlMseData = useMemo(() => {
    return {
      label: 'mse',
      data: dlResultData?.mse?.map(item => {
        return item
      }),
      type: 'line',
      backgroundColor: 'rgba(45, 199, 232, 1)',
      pointRadius: 2,
      order: 2,
    }
  }, [dlResultData?.train_loss])

  const dlThreshold = useMemo(() => {
    return {
      label: 'DL_threshold',
      type: 'line',
      borderColor: 'red',
      data: [
        {
          x: 0,
          y: dlResultData?.dl_threshold,
        },
        {
          x: dlResultData?.mse?.length,
          y: dlResultData?.dl_threshold,
        },
      ],

      order: 1,
    }
  }, [dlResultData?.dl_threshold, dlResultData?.mse?.length])

  const mseChartDataSet = useMemo(() => {
    return {
      datasets: [dlMseData, dlThreshold],
      labels: dlResultData?.mse?.map((_, i) => i),
    }
  }, [dlMseData, dlThreshold.data, dlResultData])

  const options = {
    scales: {
      x: {
        grid: {
          display: true,
          drawTicks: 1,
          color: '#383846',
        },
      },
      y: {
        beginAtZero: true,
        grid: {
          display: true,
          drawTicks: 1,
          color: '#383846',
        },
      },
    },
    responsive: true,
    maintainAspectRatio: false,
    hover: {intersect: false},
    plugins: {
      tooltip: {enabled: false},
    },
  }

  return (
    <>
      <DLNxRstChart
        isNoData={noData}
        loading={loading}
        dlResultData={dlResultData}
        trainChartDataSet={trainChartDataSet}
        mseChartDataSet={mseChartDataSet}
        options={options}
      />
    </>
  )
}

export default PredictionModalDLChart
