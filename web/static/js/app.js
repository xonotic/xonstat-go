$(document).foundation()

// Relative time conversion for all pages.
$('.abstime').each(function(i,e){
    var epoch = e.getAttribute('data-epoch');
    var d = new Date(0);
    d.setUTCSeconds(epoch);
    e.setAttribute('title', d.toDateString() + ' ' + d.toTimeString());  
});
