$(function(){
  var zero = function(s, n){
    s = '0000' + s;
    return s.substring(s.length - n);
  };
  $('.create_at').each(function(){
    var timestamp = $(this).text();
    var d = new Date();
    d.setTime(timestamp.substring(0, timestamp.length - 3));
    $(this).text(
      zero(d.getFullYear(), 4) + '-' +
      zero((d.getMonth() + 1), 2) + '-' +
      zero(d.getDate(), 2) + ' ' +
      zero(d.getHours(), 2) + ':' +
      zero(d.getMinutes(), 2) + ':' +
      zero(d.getSeconds(), 2));
    console.log(d.toString());
    $(this).show();
  });
  $('a.subject_link').click(function(){
    document.location.href = '/s/' + encodeURIComponent($(this).text());
  });
});
//  vim: set ts=2 sw=2 sts=2 expandtab fenc=utf-8:
