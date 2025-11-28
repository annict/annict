# typed: false
# frozen_string_literal: true

class FlashComponent < ApplicationComponent
  def initialize(flash)
    @flash = flash
  end

  private

  attr_reader :flash

  def flash_type
    flash.keys.first&.to_sym
  end
end
