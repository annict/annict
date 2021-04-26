# frozen_string_literal: true

module Fragment
  class StatusSelectorsController < Fragment::ApplicationController
    def index
      return head(200) unless user_signed_in?
      return head(404) unless params[:anime_ids]

      @anime_list = Anime.only_kept.where(id: params[:anime_ids].split(","))
      status_kinds = current_user.library_entries.where(work: @anime_list).status_kinds

      @anime_list.each do |anime|
        anime.status_kind = status_kinds[anime.id]
      end
    end
  end
end
