# frozen_string_literal: true

module V6::KeywordSearchable
  extend ActiveSupport::Concern

  included do
    before_action :set_search_params
  end

  private

  def set_search_params
    @search = SearchService.new(params[:q])
  end
end
