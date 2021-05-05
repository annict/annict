# frozen_string_literal: true

class FlashComponent < ApplicationComponent
  def initialize(view_context, flash:)
    super view_context
    @flash = flash
  end

  def render
    build_html do |h|
      h.tag :div,
        class: "c-flash d-none",
        data_controller: "flash",
        data_flash_type: flash_type,
        data_flash_message: @flash[flash_type] do
        h.tag :div, class: "alert alert-dismissible align-items-center border-0 d-flex mb-0 rounded-0" do
          h.tag :span, class: "c-flash__alert-icon h2 mb-0 mr-2"
          h.tag :span, class: "c-flash__message"

          h.tag :button, aria_label: "Close", class: "close", data_dismiss: "alert", type: "button" do
            h.tag :i, aria_hidden: "true", class: "far fa-times"
          end
        end
      end
    end
  end

  private

  def flash_type
    @flash.keys.first&.to_sym
  end
end
