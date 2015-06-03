class Db::EditRequestsController < ApplicationController
  def show(id)
    binding.pry
    @edit_request = EditRequest.find(id)
  end
end
