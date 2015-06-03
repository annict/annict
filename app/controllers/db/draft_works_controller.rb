class Db::DraftWorksController < Db::ApplicationController
  permits :season_id, :sc_tid, :title, :media, :official_site_url, :wikipedia_url,
          :twitter_username, :twitter_hashtag, :released_at, :released_at_about,
          edit_request_attributes: [:title, :body]

  def new
    @draft_work = DraftWork.new
    @draft_work.build_edit_request
  end

  def create(draft_work)
    @draft_work = DraftWork.new(draft_work)
    @draft_work.edit_request.user = current_user

    if @draft_work.save
      flash[:notice] = "作品の編集リクエストを作成しました"
      redirect_to db_edit_request_path(@draft_work.edit_request)
    else
      render :new
    end
  end
end
