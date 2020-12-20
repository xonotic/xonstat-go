function drawDamageChart(id, data) {
    var ctx = document.getElementById(id).getContext('2d');

    // The aspect ratio really depends on how many players are present. This formula 
    // is based on playing around with values and settling on a comfortable maximum.
    var aspectRatio = Math.min(1200/(data.distinct_players.length * 30), 7);

    window.myHorizontalBar = new Chart(ctx, {
        type: 'horizontalBar',
        data: {
            labels: data.distinct_players,
            datasets: data.damage_data,
        },
        options: {
            aspectRatio: aspectRatio,
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
                position: 'top',
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

function drawAccuracyChart(id, data) {
    var ctx = document.getElementById(id).getContext('2d');

    // The aspect ratio really depends on how many players are present. This formula 
    // is based on playing around with values and settling on a comfortable maximum.
    var aspectRatio = Math.min(1200/(data.distinct_players.length * 40), 10);

    window.myHorizontalBar = new Chart(ctx, {
        type: 'horizontalBar',
        data: {
            labels: data.distinct_players,
            datasets: data.accuracy_data,
        },
        options: {
            aspectRatio: aspectRatio,
            elements: {
                rectangle: {
                    borderWidth: 2,
                }
            },
            responsive: true,
            scales: {
                xAxes: [{
                    stacked: false,
                    gridLines: {
                        color: 'rgba(255, 255, 255, 0.2)',
                        zeroLineColor: 'rgba(255, 255, 255, 0.2)'
                    },
                    ticks: {
                        fontColor: '#ccc',
                        suggestedMin: 0,
                        suggestedMax: 100,
                        callback: function(value, index, values) {
                            return value + "%";
                        }
                    }
                }],
                yAxes: [{
                    stacked: false,
                    gridLines: {
                        color: 'rgba(255, 255, 255, 0.2)',
                        zeroLineColor: 'rgba(255, 255, 255, 0.2)'
                    },
                    ticks: {
                        fontColor: '#ccc',
                        suggestedMin: 0,
                        suggestedMax: 100,
                    }
                }]
            },
            legend: {
                position: 'top',
            },
            tooltips: {
                backgroundColor: 'rgba(255, 255, 255, 0.8)',
                titleFontColor: '#000',
                bodyFontColor: '#000',
                mode: "point",
                callbacks: {
                    label: function (tooltipItem, data) {
                        var item = data.datasets[tooltipItem.datasetIndex].richData[tooltipItem.index];
                        var accuracy = Math.round(item.pct_accuracy);
                        return `${item.weapon_cd_init_caps}: ${item.frags} frags, ${accuracy}% (${item.hit} hit/${item.fired} fired)`;

                    }
                }
            }
        }
    });
};

function drawPlayerAccuracyChart(id, data) {
    var ctx = document.getElementById(id).getContext('2d');

    window.myHorizontalBar = new Chart(ctx, {
        type: 'line',
        data: {
            labels: data.game_ids,
            datasets: data.accuracy,
        },
        options: {
            aspectRatio: 3,
            spanGaps: true,
            responsive: true,
			scales: {
                xAxes: [{
                    display: true,
                    scaleLabel: {
                        display: true,
                        labelString: 'Game #'
                    }
                }],
                yAxes: [{
                    display: true,
                    scaleLabel: {
                        display: true,
                        labelString: 'Accuracy %'
                    },
                    gridLines: {
                        color: 'rgba(255, 255, 255, 0.1)',
                        zeroLineColor: 'rgba(255, 255, 255, 0.2)'
                    },
                    ticks: {
                        suggestedMin: 0,
                    }
                }]
            },
            legend: {
                position: 'top',
            },
            tooltips: {
                mode: 'index',
                intersect: false,
            }
        }
    });
};
