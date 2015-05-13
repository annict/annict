class Db::EditRequestsController < Db::ApplicationController
  before_action :set_edit_request, only: [:show]

  private

  def set_edit_request
    @edit_request = EditRequest.find(params[:id]).decorate
  end
end
