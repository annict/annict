# frozen_string_literal: true

module V4
  class FlashComponent < V4::ApplicationComponent
    def initialize(flash)
      @flash = flash
    end

    private

    attr_reader :flash

    def flash_type
      flash.keys.first&.to_sym
    end
  end
end
