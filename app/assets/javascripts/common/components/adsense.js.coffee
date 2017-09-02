module.exports =
  template: "#t-adsence"

  props:
    adClient:
      type: String
      required: true
    adSlot:
      type: String
      required: true
    adSize:
      type: String
      required: true
    adFormat:
      type: String
      required: false
      default: "auto"
    adStyle:
      type: String
      required: false
      default: "display: block"

  mounted: ->
    (window.adsbygoogle = window.adsbygoogle || []).push({})
