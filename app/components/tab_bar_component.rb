# frozen_string_literal: true

class TabBarComponent < ApplicationComponent
  def initialize(user:)
    @user = user
  end

  private

  attr_reader :user
end
