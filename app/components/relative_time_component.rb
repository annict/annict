# frozen_string_literal: true

class RelativeTimeComponent < ApplicationComponent
  def initialize(time:, class_name: "")
    @time = time
    @class_name = class_name
  end

  private

  attr_reader :time, :class_name
end
