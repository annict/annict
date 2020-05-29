# frozen_string_literal: true

class ShareToTwitterButtonComponent < ApplicationComponent
  def initialize(text:, url:, hashtags: "", class_name: "")
    @text = text
    @url = url
    @hashtags = hashtags
    @class_name = class_name
  end

  private

  attr_reader :class_name, :text, :url, :hashtags
end
