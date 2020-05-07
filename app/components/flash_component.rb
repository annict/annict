# frozen_string_literal: true

class FlashComponent < ApplicationComponent
  def initialize(flash)
    @flash = flash
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
