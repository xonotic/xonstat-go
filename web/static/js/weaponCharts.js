function drawDamageChart(id, data) {
    var ctx = document.getElementById(id).getContext('2d');
    var isMobile = window.innerWidth < 600;

    window.damageChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: data.distinct_players,
            datasets: data.damage_data,
        },
        options: {
            indexAxis: 'y',
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                axis: 'y',
                intersect: false
            },
            layout: {
                padding: {
                    left: isMobile ? 14 : 6,
                    right: 18,
                    top: 8,
                    bottom: 8
                }
            },
            datasets: {
                bar: {
                    categoryPercentage: 0.85,
                    barPercentage: 0.9,
                    maxBarThickness: isMobile ? 32 : 38
                }
            },
            scales: {
                x: {
                    stacked: true,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)',
                    },
                    ticks: {
                        maxTicksLimit: isMobile ? 4 : 8,
                        autoSkip: true,
                        font: {
                            size: isMobile ? 10 : 12
                        }
                    }
                },
                y: {
                    stacked: true,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)',
                    },
                    ticks: {
                        autoSkip: false,
                        padding: isMobile ? 18 : 6,
                        font: {
                            size: isMobile ? 11 : 12,
                            lineHeight: isMobile ? 1.4 : 1.2
                        },
                        callback: function (value) {
                            var label = this.getLabelForValue(value);
                            if (isMobile && label && label.length > 24) {
                                return label.substring(0, 24) + '…';
                            }
                            return label;
                        }
                    }
                }
            },
            plugins: {
                legend: {
                    display: !isMobile,
                    position: 'right'
                },
                tooltip: {
                    mode: "index",
                    axis: 'y',
                    intersect: false,
                    titleColor: '#000',
                    bodyColor: '#000',
                    backgroundColor: 'rgba(255, 255, 255, 0.9)',
                    titleFont: {
                        size: isMobile ? 11 : 12,
                        lineHeight: 1.4
                    },
                    bodyFont: {
                        size: isMobile ? 11 : 12,
                        lineHeight: 1.4
                    },
                    padding: isMobile ? 10 : 6,
                    callbacks: {
                        label: function (context) {
                            var item = context.dataset.richData[context.dataIndex];
                            if (!item) {
                                return context.dataset.label || '';
                            }
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
    var isMobile = window.innerWidth < 600;

    window.accuracyChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: data.distinct_players,
            datasets: data.accuracy_data,
        },
        options: {
            indexAxis: 'y',
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'point',
                axis: 'y',
                intersect: true
            },
            datasets: {
                bar: {
                    categoryPercentage: 0.85,
                    barPercentage: 0.9,
                    maxBarThickness: isMobile ? 32 : 38
                }
            },
            layout: {
                padding: {
                    left: isMobile ? 14 : 6,
                    right: 18,
                    top: 8,
                    bottom: 8
                }
            },
            elements: {
                rectangle: {
                    borderWidth: 1,
                }
            },
            scales: {
                x: {
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)',
                    },
                    ticks: {
                        maxTicksLimit: isMobile ? 4 : 8,
                        autoSkip: true,
                        font: {
                            size: isMobile ? 10 : 12
                        },
                        callback: function (value) {
                            return value + "%";
                        }
                    }
                },
                y: {
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)',
                    },
                    ticks: {
                        autoSkip: false,
                        padding: isMobile ? 18 : 6,
                        font: {
                            size: isMobile ? 11 : 12,
                            lineHeight: isMobile ? 1.4 : 1.2
                        },
                        callback: function (value) {
                            var label = this.getLabelForValue(value);
                            if (isMobile && label && label.length > 24) {
                                return label.substring(0, 24) + '…';
                            }
                            return label;
                        }
                    }
                }
            },
            plugins: {
                legend: {
                    display: !isMobile,
                    position: 'right'
                },
                tooltip: {
                    mode: "point",
                    axis: 'y',
                    intersect: true,
                    titleColor: '#000',
                    bodyColor: '#000',
                    backgroundColor: 'rgba(255, 255, 255, 0.9)',
                    titleFont: {
                        size: isMobile ? 11 : 12,
                        lineHeight: 1.4
                    },
                    bodyFont: {
                        size: isMobile ? 11 : 12,
                        lineHeight: 1.4
                    },
                    padding: isMobile ? 10 : 6,
                    callbacks: {
                        label: function (context) {
                            var item = context.dataset.richData[context.dataIndex];
                            if (!item) {
                                return context.dataset.label || '';
                            }
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
    var isMobile = window.innerWidth < 600;

    window.playerAccuracyChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: data.game_ids,
            datasets: data.accuracy,
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            spanGaps: true,
            scales: {
                x: {
                    title: {
                        display: !isMobile,
                        text: 'Game #'
                    },
                    ticks: {
                        maxTicksLimit: isMobile ? 5 : 10,
                        font: {
                            size: isMobile ? 10 : 12
                        }
                    }
                },
                y: {
                    title: {
                        display: !isMobile,
                        text: 'Accuracy %'
                    },
                    ticks: {
                        maxTicksLimit: isMobile ? 4 : 6,
                        font: {
                            size: isMobile ? 10 : 12
                        },
                        callback: function (value) {
                            return value + "%";
                        }
                    },
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)',
                    },
                }
            },
            plugins: {
                legend: {
                    display: !isMobile,
                    position: isMobile ? 'bottom' : 'right',
                },
                tooltip: {
                    mode: "nearest",
                    intersect: true,
                    titleColor: '#000',
                    bodyColor: '#000',
                    backgroundColor: 'rgba(255, 255, 255, 0.9)',
                    titleFont: {
                        size: isMobile ? 11 : 12,
                        lineHeight: 1.4
                    },
                    bodyFont: {
                        size: isMobile ? 11 : 12,
                        lineHeight: 1.4
                    },
                    padding: isMobile ? 10 : 6,
                    callbacks: {
                        label: function (context) {
                            var accuracy = Math.round(context.dataset.data[context.dataIndex]);
                            var overallAccuracy = context.dataset.summary ? Math.round(context.dataset.summary.pct_accuracy) : accuracy;
                            var weapon = context.dataset.summary ? context.dataset.summary.weapon_cd_init_caps : context.dataset.label;
                            return `${weapon}: ${accuracy}% (${overallAccuracy}% overall)`;
                        }
                    }
                }
            },
        }
    });
};

function drawPlayerDamageChart(id, data) {
    var ctx = document.getElementById(id).getContext('2d');
    var isMobile = window.innerWidth < 600;

    window.playerDamageChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: data.game_ids,
            datasets: data.damage,
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                x: {
                    stacked: true,
                    title: {
                        display: !isMobile,
                        text: 'Game #'
                    },
                    ticks: {
                        maxTicksLimit: isMobile ? 5 : 10,
                        font: {
                            size: isMobile ? 10 : 12
                        }
                    },
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)',
                    }
                },
                y: {
                    stacked: true,
                    title: {
                        display: !isMobile,
                        text: 'Total Damage'
                    },
                    ticks: {
                        maxTicksLimit: isMobile ? 4 : 6,
                        font: {
                            size: isMobile ? 10 : 12
                        }
                    },
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)',
                    }
                }
            },
            plugins: {
                legend: {
                    display: !isMobile,
                    position: isMobile ? 'bottom' : 'right',
                },
                tooltip: {
                    mode: "index",
                    axis: 'y',
                    intersect: false,
                    titleColor: '#000',
                    bodyColor: '#000',
                    backgroundColor: 'rgba(255, 255, 255, 0.9)',
                    titleFont: {
                        size: isMobile ? 11 : 12,
                        lineHeight: 1.4
                    },
                    bodyFont: {
                        size: isMobile ? 11 : 12,
                        lineHeight: 1.4
                    },
                    padding: isMobile ? 10 : 6,
                    callbacks: {
                        label: function (context) {
                            var damage = Math.round(context.raw);
                            if (!context.dataset.summary) {
                                return damage + " damage";
                            }
                            return context.dataset.summary.weapon_cd_init_caps + ": " + damage + " damage";
                        }
                    }
                }
            }
        }
    });
};