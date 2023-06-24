function drawDamageChart(id, data) {
    var ctx = document.getElementById(id).getContext('2d');

    // The aspect ratio really depends on how many players are present. This formula 
    // is based on playing around with values and settling on a comfortable maximum.
    var barHeight = 30;
    var barPadding = 5;
    var aspectRatio = Math.min(1200 / (data.distinct_players.length * (barHeight + barPadding)), 7);

    window.damageChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: data.distinct_players,
            datasets: data.damage_data,
        },
        options: {
            indexAxis: 'y',
            aspectRatio: aspectRatio,
            scales: {
                x: {
                    stacked: true,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)',
                    }
                },
                y: {
                    stacked: true,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)',
                    }
                }
            },
            plugins: {
                tooltip: {
                    mode: "point",
                    titleColor: '#000',
                    bodyColor: '#000',
                    backgroundColor: 'rgba(255, 255, 255, 0.8)',
                    callbacks: {
                        label: function (context) {
                            console.log(context);
                            // var item = data.datasets[tooltipItem.datasetIndex].richData[tooltipItem.index];
                            var item = context.dataset.richData[context.dataIndex];
                            var damage = Math.round(item.pct_total_damage);
                            return `${item.weapon_cd_init_caps}: ${item.frags} frags, ${item.actual} damage (${damage}% of total)`;

                        }
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
    var barHeight = 40;
    var barPadding = 5;
    var aspectRatio = Math.min(1200 / (data.distinct_players.length * (barHeight + barPadding)), 10);

    window.accuracyChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: data.distinct_players,
            datasets: data.accuracy_data,
        },
        options: {
            indexAxis: 'y',
            aspectRatio: aspectRatio,
            elements: {
                rectangle: {
                    borderWidth: 2,
                }
            },
            scales: {
                x: {
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)',
                    },
                    ticks: {
                        callback: function (value, index, values) {
                            return value + "%";
                        }
                    }
                },
                y: {
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)',
                    },
                }
            },
            plugins: {
                tooltip: {
                    mode: "point",
                    titleColor: '#000',
                    bodyColor: '#000',
                    backgroundColor: 'rgba(255, 255, 255, 0.8)',
                    callbacks: {
                        label: function (context) {
                            var item = context.dataset.richData[context.dataIndex];
                            var accuracy = Math.round(item.pct_accuracy);
                            return `${item.weapon_cd_init_caps}: ${item.frags} frags, ${accuracy}% (${item.hit} hit/${item.fired} fired)`;
                        }
                    }
                }
            }
        }
    });
};

function drawPlayerAccuracyChart(id, data) {
    var ctx = document.getElementById(id).getContext('2d');

    window.playerAccuracyChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: data.game_ids,
            datasets: data.accuracy,
        },
        options: {
            aspectRatio: 3,
            spanGaps: true,
            scales: {
                x: {
                    title: {
                        display: true,
                        text: 'Game #'
                    },
                },
                y: {
                    title: {
                        display: true,
                        text: 'Accuracy %'
                    },
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)',
                    },
                }
            },
            plugins: {
                tooltip: {
                    mode: "point",
                    titleColor: '#000',
                    bodyColor: '#000',
                    backgroundColor: 'rgba(255, 255, 255, 0.8)',
                    callbacks: {
                        label: function (context) {
                            var accuracy = Math.round(context.dataset.data[context.dataIndex]);
                            var overallAccuracy = Math.round(context.dataset.summary.pct_accuracy);
                            return `${context.dataset.summary.weapon_cd_init_caps}: ${accuracy}% (${overallAccuracy}% overall)`;
                        }
                    }
                }
            },
        }
    });
};

function drawPlayerDamageChart(id, data) {
    var ctx = document.getElementById(id).getContext('2d');

    window.playerDamageChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: data.game_ids,
            datasets: data.damage,
        },
        options: {
            aspectRatio: 3,
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
                }],
                yAxes: [{
                    stacked: true,
                    gridLines: {
                        color: 'rgba(255, 255, 255, 0.2)',
                        zeroLineColor: 'rgba(255, 255, 255, 0.2)'
                    },
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
                    label: function (tooltipItem, d) {
                        var label = d.datasets[tooltipItem.datasetIndex].label;
                        var totalDamage = data.total_damage_per_game[tooltipItem.index];
                        var weaponDamage = d.datasets[tooltipItem.datasetIndex].data[tooltipItem.index];
                        var pctOfTotal = Math.round(weaponDamage/totalDamage*100);

                        return `${label}: ${weaponDamage} (${pctOfTotal}% of total)`;
                    }
                }
            }
        }
    });
};