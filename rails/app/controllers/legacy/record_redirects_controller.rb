# typed: false
# frozen_string_literal: true

module Legacy
  class RecordRedirectsController < ApplicationController
    def show
      url = case params[:provider]
      when "tw"
        EpisodeRecord.only_kept.find_by!(twitter_url_hash: params[:url_hash]).share_url_with_query(:twitter)
      when "fb"
        EpisodeRecord.only_kept.find_by!(facebook_url_hash: params[:url_hash]).share_url_with_query(:facebook)
      else
        root_path
      end

      redirect_to url, status: 301
    end
  end
end
