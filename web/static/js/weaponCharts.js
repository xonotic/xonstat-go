// Colors assigned to the various weapons
var weaponColors = {
    "arc": "#7c9ceb",
    "laser": "#f7717b",
    "blaster": "#f7717b",
    "shotgun": "#849ba8",
    "uzi": "#81f13d",
    "machinegun": "#81f13d",
    "grenadelauncher": "#fd7865",
    "mortar": "#fd7865",
    "minelayer": "#fd7865",
    "electro": "#6899f2",
    "crylink": "#ea6ff9",
    "nex": "#75c3d5",
    "vortex": "#75c3d5",
    "hagar": "#e39160",
    "rocketlauncher": "#e9be57",
    "devastator": "#e9be57",
    "porto": "#6899f2",
    "minstanex": "#978ed2",
    "vaporizer": "#978ed2",
    "hook": "#81f13d",
    "hlac": "#e5965b",
    "seeker": "#f7717b",
    "rifle": "#e39160",
    "tuba": "#e9be57",
    "fireball": "#f0855f"
};

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
                }],
                yAxes: [{
                    stacked: true,
                }]
            },
            legend: {
                position: 'right',
            },
            tooltips: {
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
