# typed: false
# frozen_string_literal: true

module KeywordSearchable
  extend ActiveSupport::Concern

  included do
    before_action :set_search_params
  end

  private

  def set_search_params
    @search = Deprecated::SearchService.new(params[:q])
  end
end
