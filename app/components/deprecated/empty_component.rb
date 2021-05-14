# frozen_string_literal: true

module Deprecated
  class EmptyComponent < Deprecated::ApplicationComponent
    def initialize(text:)
      @text = text
    end

    private

    attr_reader :text
  end
end
