(function ($) {
    $.fn.serializeFormJSON = function () {

        var o = {};
        var a = this.serializeArray();
        $.each(a, function () {
            if (o[this.name]) {
                if (!o[this.name].push) {
                    o[this.name] = [o[this.name]];
                }
                o[this.name].push(this.value || '');
            } else {
                o[this.name] = this.value || '';
            }
        });
        return o;
    };
})(jQuery);


$(function () {
    console.log('form started');
    $(".hide-me").hide();
    $("form").on("submit", function (e) {
        e.preventDefault();

        console.log("form submitting");

        let form = $(this);
        $.ajax({
            type: "post",
            url: "http://dm-demo-default.apps.cluster-anz-f723.anz-f723.openshiftworkshop.com/form",
            // url: "http://localhost:10000/form",
            contentType: "application/json",
            data: JSON.stringify(form.serializeFormJSON()),
            success: function (data) {
                $("#_firstNameValid").text("First Name Valid: " + data.validFirstName);
                console.log(data.message)
                $("#_lastNameValid").text("Last Name Valid: " + data.validLastName);
                $("#_abnStatus").text("ABN Status: " + data.abnStatus);
                $("#_message").text("Message: " + data.message);
                $(".hide-me").show();
            }
        });
    });
});