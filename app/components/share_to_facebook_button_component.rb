# frozen_string_literal: true

# TODO: Buttons::ShareToFacebookButtonComponent に移行する
class ShareToFacebookButtonComponent < ApplicationComponent
  def initialize(url:, class_name: "")
    @url = url
    @class_name = class_name
  end

  private

  attr_reader :class_name, :url
end
