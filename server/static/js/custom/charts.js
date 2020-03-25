class MyChart {
    constructor(name, ext_id, value, visualisation) {
        this.name = name
        this.ext_id = ext_id
        this.data = JSON.parse(value)
        this.visualisation = visualisation
        this.keys = Object.keys(this.data)
    }

    generateChart() {
        let backgroundColors = []
        let borderColors = []
        for(let i = 0; i < this.keys.length; i++) {
            let [bg, border] = this.generateColor()
            backgroundColors.push(bg)
            borderColors.push(border)
        }

        let dataset = {
            label: this.name,
            data: Object.values(this.data),
            backgroundColor: backgroundColors,
            borderColor: borderColors,
            borderWidth: 1
        }

        let chartData = {
            type: this.visualisation,
            data: {
                labels: this.keys,
                datasets: [dataset]
            },
            options: {
                responsive: true,
                legend: false,
                layout: {
                    padding: {
                        left: 20,
                        right: 0,
                        top: 0,
                        bottom: 0
                    }
                },
                tooltips: {
                    callbacks: {
                        label: function(item) {
                            return item.yLabel + "€ per day"
                        }
                    }
                },
                scales: {
                    yAxes: [{
                        ticks: {
                            beginAtZero: true
                        }
                    }],
                }
            }
        }

        return chartData
    }

    generateColor() {
        let raw_red = Math.floor(Math.random() * 255)
        let raw_green = Math.floor(Math.random() * 255)
        let raw_blue = Math.floor(Math.random() * 255)

        const [red, green, blue] = this.preventWhite(raw_red, raw_green, raw_blue)

        let backgroundColor = "rgba("+red+","+green+","+blue+",0.4)"
        let borderColor = "rgba("+red+","+green+","+blue+",1)"

        return [backgroundColor, borderColor]
    }

    preventWhite(red, green, blue) {
        if (red > 180 && green > 180 && blue > 180) {
            let i = Math.floor(Math.random() * 2)
            if (i == 0) {
                red -= Math.floor(Math.random() * 180)
            } else if(i == 1) {
                green -= Math.floor(Math.random() * 180)
            } else if(i == 2) {
                blue -= Math.floor(Math.random() * 180)
            }
        }
        return [red, green, blue]
    }
}

$(document).ready(function() {
    let chartObjs = $(".dataChart")

    chartObjs.each(function(index) {
        let name = $(this).find(".name").text()
        let extID = $(this).find(".extID").text()
        let value = $(this).find(".value").text()
        let visualisation = $(this).find(".visualisation").text()

        $(this).find(".info").remove()

        let chart = new MyChart(name, extID, value, visualisation)
        let data = chart.generateChart()

        let canvas = $(this).find(".chartjs_graph")
        new Chart(canvas, data)
    })
    // for(let obj in chartObjs) {
    //     console.log(obj)
    // }
})