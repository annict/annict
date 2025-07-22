# typed: false
# frozen_string_literal: true

RSpec.describe "GET /fragment/works/:work_id/collection_items/new", type: :request do
  it "未ログインの場合、ログインページにリダイレクトされること" do
    work = FactoryBot.create(:work)

    get fragment_new_collection_item_path(work)

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログイン済みで、コレクションが存在しない場合、フォームが無効化されて表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    get fragment_new_collection_item_path(work)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("turbo-frame id=\"c-new-collection-item\"")
    expect(response.body).to include("disabled=\"disabled\"")
    expect(response.body).to include(I18n.t("messages.fragment.collection_items.new.disabled_form_hint1"))
  end

  it "ログイン済みで、選択可能なコレクションが存在する場合、フォームが有効化されて表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user:)
    login_as(user, scope: :user)

    get fragment_new_collection_item_path(work)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("turbo-frame id=\"c-new-collection-item\"")
    expect(response.body).not_to include("disabled=\"disabled\"")
    expect(response.body).to include(collection.name)
  end

  it "ログイン済みで、作品が既にコレクションに追加されている場合、そのコレクションは選択肢に表示されず、追加済みとして表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection1 = FactoryBot.create(:collection, user:, name: "お気に入り")
    collection2 = FactoryBot.create(:collection, user:, name: "後で見る")
    FactoryBot.create(:collection_item, collection: collection1, work:, user:)
    login_as(user, scope: :user)

    get fragment_new_collection_item_path(work)

    expect(response).to have_http_status(:ok)
    # collection1は追加済みとして表示される
    expect(response.body).to include("<a class=\"badge bg-secondary text-white\"")
    expect(response.body).to include("お気に入り")
    # collection2は選択肢として表示される
    expect(response.body).to include("<option value=\"#{collection2.id}\">後で見る</option>")
    # collection1は選択肢には表示されない
    expect(response.body).not_to include("<option value=\"#{collection1.id}\">お気に入り</option>")
  end

  it "削除された作品の場合、404エラーが返されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, deleted_at: Time.current)
    login_as(user, scope: :user)

    expect {
      get fragment_new_collection_item_path(work)
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
