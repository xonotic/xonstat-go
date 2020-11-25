function drawDamageChart(id, data) {
    var ctx = document.getElementById('damageChart').getContext('2d');
    window.myHorizontalBar = new Chart(ctx, {
        type: 'horizontalBar',
        data: {
            labels: data.distinct_players,
            datasets: data.damage_data,
        },
        options: {
            elements: {
                rectangle: {
                    borderWidth: 2,
                }
            },
            responsive: true,
            scales: {
                xAxes: [{
                    stacked: true,
                    gridLines: {
                        color: 'rgba(255, 255, 255, 0.2)',
                        zeroLineColor: 'rgba(255, 255, 255, 0.2)'
                    },
                    ticks: {
                        fontColor: '#ccc'
                    }
                }],
                yAxes: [{
                    stacked: true,
                    gridLines: {
                        color: 'rgba(255, 255, 255, 0.2)',
                        zeroLineColor: 'rgba(255, 255, 255, 0.2)'
                    },
                    ticks: {
                        fontColor: '#ccc'
                    }
                }]
            },
            legend: {
                position: 'right',
            },
            tooltips: {
                backgroundColor: 'rgba(255, 255, 255, 0.8)',
                titleFontColor: '#000',
                bodyFontColor: '#000',
                mode: "point",
                callbacks: {
                    label: function (tooltipItem, data) {
                        var item = data.datasets[tooltipItem.datasetIndex].richData[tooltipItem.index];
                        var damage = Math.round(item.pct_total_damage);
                        return `${item.weapon_cd_init_caps}: ${item.frags} frags, ${item.actual} damage (${damage}% of total)`;

                    }
                }
            }
        }
    });
};
