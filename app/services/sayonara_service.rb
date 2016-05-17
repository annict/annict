# frozen_string_literal: true

class SayonaraService
  def initialize(user)
    @user = user
  end

  def bye_bye!
    ActiveRecord::Base.transaction do
      @user.destroy
      @user.oauth_applications.available.find_each do |app|
        app.update(owner: nil)
        app.hide!
      end
    end
  end
end
