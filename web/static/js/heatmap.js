d3.json("/heatmap", {
    headers: new Headers({
        "Accept": "application/json"
    }),
}).then((data) => {

    var gameCounts = data.map(d => d[2]);
    var minGames = Math.min(...gameCounts);
    var maxGames = Math.max(...gameCounts);

    function makeSparse(data) {
        var sparse = [];
        for (day = 0; day < 7; day++) {
            for (hour = 0; hour < 24; hour++) {
                sparse[(day * 24) + hour] = [day, hour, 0]
            }
        }

        data.forEach(element => {
            dayOfWeek = element[0];
            hourOfDay = element[1];
            gameCount = element[2];

            i = (dayOfWeek * 24) + hourOfDay;
            sparse[i] = [dayOfWeek, hourOfDay, gameCount];
        });

        return sparse;
    }

    data = makeSparse(data);

    // set the dimensions and margins of the graph
    var margin = { top: 40, right: 25, bottom: 30, left: 80 },
        width = 900 - margin.left - margin.right,
        height = 450 - margin.top - margin.bottom;

    // append the svg object to the body of the page
    var svg = d3.select("#heatmap")
        .append("svg")
        .attr("width", width + margin.left + margin.right)
        .attr("height", height + margin.top + margin.bottom)
        .append("g")
        .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    // Labels of row and columns
    var dayLabels = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"];

    var hourLabels = [
        "12AM", "1AM", "2AM", "3AM", "4AM", "5AM", "6AM", "7AM", "8AM", "9AM", "10AM", "11AM",
        "12PM", "1PM", "2PM", "3PM", "4PM", "5PM", "6PM", "7PM", "8PM", "9PM", "10PM", "11PM"
    ];

    // Build X scales and axis:
    var x = d3.scaleBand()
        .range([0, width])
        .domain(hourLabels)
        .padding(0.05);

    svg.append("g")
        .style("font-size", 15)
        .attr("transform", "translate(0," + height + ")")
        .call(d3.axisBottom(x).tickSize(0).ticks(24).tickValues(hourLabels))
        .select(".domain").remove()

    // Build Y scales and axis:
    var y = d3.scaleBand()
        .range([height, 0])
        .domain(dayLabels)
        .padding(0.05);

    svg.append("g")
        .style("font-size", 15)
        .call(d3.axisLeft(y).ticks(7).tickSize(0).tickValues(dayLabels))
        .select(".domain").remove()

    // Build color scale
    var colorScale = d3.scaleSequential()
        .interpolator(d3.interpolateOranges)
        .domain([0, maxGames])

    // Tooltip
    svg.append("text")
        .attr("id", "tooltip")
        .attr("x", 0)
        .attr("y", -10)
        .attr("text-anchor", "left")
        .style("stroke", "#08C")
        .style("fill", "#08C")
        .style("font-size", "22px")
        .style("font-family", "xolonium");

    // Three function that change the tooltip when user hover / move / leave a cell
    var mouseover = function (event, d) {
        d3.select(this)
            .style("stroke", "#444")
            .style("opacity", 0.8)
    }
    var mousemove = function (event, d) {
        d3.select("#tooltip").text(`${dayLabels[d[0]-1]} ${hourLabels[d[1]]}: ${d[2]} games`);
    }

    var mouseleave = function (event, d) {
        d3.select("#tooltip").text("");
        d3.select(this)
            .style("stroke", "none")
            .style("opacity", 0.8)
    }

    // add the squares
    svg.selectAll()
        .data(data, function (d) { return d[2]; })
        .enter()
        .append("rect")
        .attr("x", function (d) { return x(hourLabels[d[1]]) })
        .attr("y", function (d) { return y(dayLabels[d[0]-1]) })
        .attr("rx", 4)
        .attr("ry", 4)
        .attr("width", x.bandwidth())
        .attr("height", y.bandwidth())
        .style("fill", function (d) { return colorScale(d[2]) })
        .style("stroke-width", 4)
        .style("stroke", "none")
        .style("opacity", 0.8)
        .on("mouseover", mouseover)
        .on("mousemove", mousemove)
        .on("mouseleave", mouseleave)

})