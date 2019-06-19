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
            url: "http://currencyvalidation-currency-rules.apps.cluster-anz-f723.anz-f723.openshiftworkshop.com/camel/currency/validate",
            contentType: "application/json",
            data: JSON.stringify(form.serializeFormJSON()),
            success: function (data) {
                console.log(data.message)
                $("#_message").text("Currency code is valid: " + data.valid);
                $(".hide-me").show();
            }
        });
    });
});
