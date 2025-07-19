# typed: false
# frozen_string_literal: true

RSpec.describe "GET /collection_items/:collection_item_id/edit", type: :request do
  it "ログインしているとき、自分のコレクションアイテムの編集ページが表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)
    collection_item = FactoryBot.create(:collection_item, user: user, work: work, collection: collection)

    login_as(user, scope: :user)

    get "/collection_items/#{collection_item.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
    expect(response.body).to include(collection.name)
  end

  it "ログインしていないとき、ログインページにリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)
    collection_item = FactoryBot.create(:collection_item, user: user, work: work, collection: collection)

    get "/collection_items/#{collection_item.id}/edit"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "他のユーザーのコレクションアイテムを編集しようとしたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user)
    other_user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    other_collection = FactoryBot.create(:collection, user: other_user)
    other_collection_item = FactoryBot.create(:collection_item, user: other_user, work: work, collection: other_collection)

    login_as(user, scope: :user)

    expect {
      get "/collection_items/#{other_collection_item.id}/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないコレクションアイテムを編集しようとしたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user)

    login_as(user, scope: :user)

    expect {
      get "/collection_items/999999/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済みのコレクションアイテムを編集しようとしたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)
    collection_item = FactoryBot.create(:collection_item, user: user, work: work, collection: collection)
    collection_item.destroy!

    login_as(user, scope: :user)

    expect {
      get "/collection_items/#{collection_item.id}/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
