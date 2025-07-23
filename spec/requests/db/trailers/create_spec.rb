# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/works/:work_id/trailers", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = create(:work)
    form_params = {
      rows: "https://www.youtube.com/watch?v=nGgm5yBznTM,第1弾"
    }

    post "/db/works/#{work.id}/trailers", params: {deprecated_db_trailer_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Trailer.all.size).to eq(0)
  end

  it "エディター権限を持たないユーザーでログインしているとき、アクセスできないこと" do
    work = create(:work)
    user = create(:registered_user)
    form_params = {
      rows: "https://www.youtube.com/watch?v=nGgm5yBznTM,第1弾"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/trailers", params: {deprecated_db_trailer_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Trailer.all.size).to eq(0)
  end

  it "エディター権限を持つユーザーでログインしているとき、トレイラーを作成できること" do
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "https://www.youtube.com/watch?v=nGgm5yBznTM,第1弾"
    }

    # YouTubeサムネイルの取得をスタブ化
    allow(HTTParty).to receive(:get).and_return(instance_double(HTTParty::Response, code: 200))
    allow(Down).to receive(:open).and_return(instance_double(Tempfile))
    # Trailer#saveのフックで呼ばれるimage=メソッドをスキップ
    allow(Trailer).to receive(:new).and_wrap_original do |method, *args|
      trailer = method.call(*args)
      allow(trailer).to receive(:image=)
      trailer
    end

    expect(Trailer.all.size).to eq(0)

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/trailers", params: {deprecated_db_trailer_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Trailer.all.size).to eq(1)

    trailer = Trailer.last
    expect(trailer.title).to eq("第1弾")
    expect(trailer.url).to eq("https://www.youtube.com/watch?v=nGgm5yBznTM")
    expect(trailer.work_id).to eq(work.id)
  end

  it "エディター権限を持つユーザーでログインしているとき、複数のトレイラーを一度に作成できること" do
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "https://www.youtube.com/watch?v=aaa,第1弾\nhttps://www.youtube.com/watch?v=bbb,第2弾"
    }

    # YouTubeサムネイルの取得をスタブ化
    allow(HTTParty).to receive(:get).and_return(instance_double(HTTParty::Response, code: 200))
    allow(Down).to receive(:open).and_return(instance_double(Tempfile))
    # Trailer#saveのフックで呼ばれるimage=メソッドをスキップ
    allow(Trailer).to receive(:new).and_wrap_original do |method, *args|
      trailer = method.call(*args)
      allow(trailer).to receive(:image=)
      trailer
    end

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/trailers", params: {deprecated_db_trailer_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Trailer.all.size).to eq(2)

    trailers = Trailer.order(:sort_number)
    expect(trailers[0].title).to eq("第1弾")
    expect(trailers[0].url).to eq("https://www.youtube.com/watch?v=aaa")
    expect(trailers[1].title).to eq("第2弾")
    expect(trailers[1].url).to eq("https://www.youtube.com/watch?v=bbb")
  end

  it "エディター権限を持つユーザーでログインしているとき、バリデーションエラーがある場合はnewページが表示されること" do
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: ""
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/trailers", params: {deprecated_db_trailer_rows_form: form_params}

    expect(response.status).to eq(422)
    # render_templateは使えないので、別の方法でチェック
    expect(response.body).to include("PVを登録する")
    expect(Trailer.all.size).to eq(0)
  end

  it "存在しない作品IDを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "https://www.youtube.com/watch?v=nGgm5yBznTM,第1弾"
    }

    login_as(user, scope: :user)

    expect {
      post "/db/works/99999/trailers", params: {deprecated_db_trailer_rows_form: form_params}
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
