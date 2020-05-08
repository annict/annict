# frozen_string_literal: true

class ProfileHeaderComponent < ApplicationComponent
  def initialize(user_entity:, current_user:)
    @user_entity = user_entity
    @current_user = current_user
  end

  private

  attr_reader :user_entity, :current_user
end
