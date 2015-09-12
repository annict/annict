class ProvidersController < ApplicationController
  before_filter :authenticate_user!

  def destroy(id)
    current_user.providers.destroy(id)
    redirect_to :back, notice: "連携を解除しました"
  end
end
