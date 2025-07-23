# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/works/:work_id/episodes", type: :request do
  it "ログインしていない場合、ログインページにリダイレクトされること" do
    work = FactoryBot.create(:work)
    form_params = {
      rows: "#1,1,The episode"
    }

    post "/db/works/#{work.id}/episodes", params: {deprecated_db_episode_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Episode.all.size).to eq(0)
  end

  it "エディター権限を持たないユーザーがログインしている場合、アクセスできないこと" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user)
    form_params = {
      rows: "#1,1,The episode"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/episodes", params: {deprecated_db_episode_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Episode.all.size).to eq(0)
  end

  it "エディター権限を持つユーザーがログインしている場合、エピソードを作成できること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    form_params = {
      rows: "第127話,127,逆転！稲妻の戦士\r\n第128話,128,城之内 死す"
    }

    login_as(user, scope: :user)

    expect(Episode.all.size).to eq(0)

    post "/db/works/#{work.id}/episodes", params: {deprecated_db_episode_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")

    expect(Episode.all.size).to eq(2)
    episode = Episode.last

    expect(episode.number).to eq("第128話")
    expect(episode.raw_number).to eq(128)
    expect(episode.title).to eq("城之内 死す")
  end

  it "エディター権限を持つユーザーがログインしている場合、rowsが空のときバリデーションエラーになること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    form_params = {
      rows: ""
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/episodes", params: {deprecated_db_episode_rows_form: form_params}

    expect(response.status).to eq(422)
    expect(Episode.all.size).to eq(0)
  end

  it "エディター権限を持つユーザーがログインしている場合、rowsがnilのときバリデーションエラーになること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    form_params = {
      rows: nil
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/episodes", params: {deprecated_db_episode_rows_form: form_params}

    expect(response.status).to eq(422)
    expect(Episode.all.size).to eq(0)
  end

  it "エディター権限を持つユーザーがログインしている場合、カンマのない不正なフォーマットでもエピソードを作成できること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    form_params = {
      rows: "第1話"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/episodes", params: {deprecated_db_episode_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")

    expect(Episode.all.size).to eq(1)
    episode = Episode.last

    expect(episode.number).to eq("第1話")
    expect(episode.raw_number).to be_nil
    expect(episode.title).to be_nil
  end

  it "エディター権限を持つユーザーがログインしている場合、ダブルクオートを含むタイトルのエピソードを作成できること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    form_params = {
      rows: "第1話,1,「最初」の物語"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/episodes", params: {deprecated_db_episode_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")

    expect(Episode.all.size).to eq(1)
    episode = Episode.last

    expect(episode.number).to eq("第1話")
    expect(episode.raw_number).to eq(1)
    expect(episode.title).to eq("「最初」の物語")
  end

  it "エディター権限を持つユーザーがログインしている場合、前後に空白があってもトリムされてエピソードを作成できること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    form_params = {
      rows: "  第1話  ,  1  ,  最初の物語  "
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/episodes", params: {deprecated_db_episode_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")

    expect(Episode.all.size).to eq(1)
    episode = Episode.last

    expect(episode.number).to eq("第1話")
    expect(episode.raw_number).to eq(1)
    expect(episode.title).to eq("最初の物語")
  end

  it "エディター権限を持つユーザーがログインしている場合、空行を含むrowsでも空行を無視してエピソードを作成できること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    form_params = {
      rows: "第1話,1,最初の物語\r\n\r\n第2話,2,次の物語"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/episodes", params: {deprecated_db_episode_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")

    expect(Episode.all.size).to eq(2)

    first_episode = Episode.first
    expect(first_episode.number).to eq("第1話")
    expect(first_episode.raw_number).to eq(1)
    expect(first_episode.title).to eq("最初の物語")

    last_episode = Episode.last
    expect(last_episode.number).to eq("第2話")
    expect(last_episode.raw_number).to eq(2)
    expect(last_episode.title).to eq("次の物語")
  end

  it "エディター権限を持つユーザーがログインしている場合、sort_numberが正しく設定されること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)

    # 既存のエピソードを作成
    FactoryBot.create(:episode, work: work, sort_number: 100)

    form_params = {
      rows: "第2話,2,次の物語\r\n第3話,3,最後の物語"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/episodes", params: {deprecated_db_episode_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")

    episodes = Episode.where(work: work).order(:sort_number)
    expect(episodes.size).to eq(3)

    expect(episodes[0].sort_number).to eq(100)
    expect(episodes[1].sort_number).to eq(200)
    expect(episodes[2].sort_number).to eq(300)
  end

  it "エディター権限を持つユーザーがログインしている場合、prev_episode_idが正しく設定されること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    form_params = {
      rows: "第1話,1,最初の物語"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/episodes", params: {deprecated_db_episode_rows_form: form_params}

    expect(response.status).to eq(302)

    episode = Episode.last
    expect(episode.prev_episode_id).to be_nil

    # 2つ目のエピソードを追加
    form_params2 = {
      rows: "第2話,2,次の物語"
    }

    post "/db/works/#{work.id}/episodes", params: {deprecated_db_episode_rows_form: form_params2}

    episode2 = Episode.last
    expect(episode2.prev_episode_id).to eq(episode.id)
  end
end
