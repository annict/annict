# frozen_string_literal: true

module V6::KeywordSearchable
  extend ActiveSupport::Concern

  private

  def set_search_params
    @search = SearchService.new(params[:q])
  end
end
