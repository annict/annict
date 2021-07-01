# frozen_string_literal: true

class EmptyComponent < ApplicationComponent
  def initialize(text:)
    @text = text
  end

  private

  attr_reader :text
end
