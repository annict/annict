# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /collection_items/:collection_item_id", type: :request do
  it "ログインしているとき、コレクションアイテムが削除されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)
    collection_item = FactoryBot.create(:collection_item, user: user, work: work, collection: collection)

    login_as(user, scope: :user)

    expect(CollectionItem.count).to eq(1)

    delete "/collection_items/#{collection_item.id}"

    expect(CollectionItem.count).to eq(0)
    expect(response.status).to eq(302)
    expect(response).to redirect_to(user_collection_path(user.username, collection.id))
    expect(flash[:notice]).to eq(I18n.t("messages._common.deleted"))
  end

  it "ログインしていないとき、ログインページにリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)
    collection_item = FactoryBot.create(:collection_item, user: user, work: work, collection: collection)

    delete "/collection_items/#{collection_item.id}"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "他のユーザーのコレクションアイテムを削除しようとしたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user)
    other_user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    other_collection = FactoryBot.create(:collection, user: other_user)
    other_collection_item = FactoryBot.create(:collection_item, user: other_user, work: work, collection: other_collection)

    login_as(user, scope: :user)

    expect {
      delete "/collection_items/#{other_collection_item.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないコレクションアイテムを削除しようとしたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user)

    login_as(user, scope: :user)

    delete "/collection_items/999999"

    expect(response.status).to eq(404)
  end

  it "削除済みのコレクションアイテムを削除しようとしたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user: user)
    collection_item = FactoryBot.create(:collection_item, user: user, work: work, collection: collection)
    collection_item.destroy!

    login_as(user, scope: :user)

    expect {
      delete "/collection_items/#{collection_item.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
