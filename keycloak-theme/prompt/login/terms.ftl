<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=false; section>
    <#if section = "header">
        Datenschutzerklärung &amp; Einwilligung
    <#elseif section = "show-username">
        <link rel="stylesheet" href="${url.resourcesPath}/css/privacy-consent.css"/>
    <#elseif section = "form">
        <div id="kc-terms-text">
            <#include "privacy-content.html">
        </div>
        <form class="form-actions terms-actions" action="${url.loginAction}" method="POST">
            <input class="${properties.kcButtonClass!} ${properties.kcButtonPrimaryClass!} ${properties.kcButtonLargeClass!}"
                   name="accept" id="kc-accept" type="submit" value="${msg("doAccept")}"/>
            <input class="${properties.kcButtonClass!} ${properties.kcButtonDefaultClass!} ${properties.kcButtonLargeClass!}"
                   name="cancel" id="kc-decline" type="submit" value="${msg("doDecline")}"/>
        </form>
    </#if>
</@layout.registrationLayout>
