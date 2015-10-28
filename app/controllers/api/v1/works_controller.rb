class Api::V1::WorksController < Api::V1::ApplicationController
  def index
    @works = Work.limit(5)
  end
end
