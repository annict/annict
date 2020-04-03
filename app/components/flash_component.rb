# frozen_string_literal: true

class FlashComponent < ApplicationComponent
  def initialize(flash)
    @flash = flash
  end

  def call
    return unless flash_key

    Htmlrb.build do |el|
      el.div class: "alert alert-dismissible align-content-center d-flex mb-0 #{alert_class}" do
        el.i(class: "far h2 mb-0 mr-2 #{alert_icon_class}") {}

        el.span do
          flash[flash_key]
        end

        el.button aria_label: "Close", class: "close", data_dismiss: "alert", type: "button" do
          el.i(aria_hidden: "true", class: "fas fa-times") {}
        end
      end
    end.html_safe
  end

  private

  attr_reader :flash

  def alert_class
    case flash_key
    when :notice then "alert-success"
    when :alert then "alert-danger"
    end
  end

  def alert_icon_class
    case flash_key
    when :notice then "fa-check-circle"
    when :alert then "fa-exclamation-triangle"
    end
  end

  def flash_key
    flash.keys.first&.to_sym
  end
end
