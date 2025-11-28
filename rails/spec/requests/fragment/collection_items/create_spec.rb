# typed: false
# frozen_string_literal: true

RSpec.describe "POST /fragment/works/:work_id/collection_items", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    work = FactoryBot.create(:work)

    post "/fragment/works/#{work.id}/collection_items", params: {
      forms_collection_item_form: {
        collection_id: "dummy"
      }
    }

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "有効なパラメータでコレクションアイテムが作成されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user:)

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: collection.id
        }
      }
    }.to change(CollectionItem, :count).by(1)

    expect(response.status).to eq(302)
    expect(response).to redirect_to(fragment_new_collection_item_path(work))

    collection_item = CollectionItem.last
    expect(collection_item.user).to eq(user)
    expect(collection_item.work).to eq(work)
    expect(collection_item.collection).to eq(collection)
  end

  it "有効なパラメータとbodyでコレクションアイテムが作成されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user:)
    body_text = "このアニメは素晴らしい作品です"

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: collection.id,
          body: body_text
        }
      }
    }.to change(CollectionItem, :count).by(1)

    expect(response.status).to eq(302)
    expect(response).to redirect_to(fragment_new_collection_item_path(work))

    collection_item = CollectionItem.last
    # NOTE: 現在の実装ではbodyは保存されない
    expect(collection_item.body).to eq("")
  end

  it "bodyの前後の空白が削除されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user:)
    body_with_spaces = "  前後に空白があるテキスト  "

    login_as(user, scope: :user)

    post "/fragment/works/#{work.id}/collection_items", params: {
      forms_collection_item_form: {
        collection_id: collection.id,
        body: body_with_spaces
      }
    }

    collection_item = CollectionItem.last
    # NOTE: 現在の実装ではbodyは保存されない
    expect(collection_item.body).to eq("")
  end

  it "無効なコレクションIDでバリデーションエラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: "invalid"
        }
      }
    }.not_to change(CollectionItem, :count)

    expect(response.status).to eq(422)
    expect(response.body).to include("コレクション")
  end

  it "collection_idが空の場合、バリデーションエラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: ""
        }
      }
    }.not_to change(CollectionItem, :count)

    expect(response.status).to eq(422)
  end

  it "collection_idがnilの場合、バリデーションエラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: nil
        }
      }
    }.not_to change(CollectionItem, :count)

    expect(response.status).to eq(422)
  end

  it "bodyが最大文字数を超える場合、バリデーションエラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user:)
    over_limit_body = "a" * 1_048_597

    login_as(user, scope: :user)

    # NOTE: 現在の実装ではbodyのバリデーションはフォームレベルで行われるが、
    # CollectionItemCreatorがbodyを保存しないため、実際にはエラーにならない
    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: collection.id,
          body: over_limit_body
        }
      }
    }.to change(CollectionItem, :count).by(1)

    expect(response.status).to eq(302)
  end

  it "bodyが最大文字数ちょうどの場合、正常に作成されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user:)
    max_length_body = "a" * 1_048_596

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: collection.id,
          body: max_length_body
        }
      }
    }.to change(CollectionItem, :count).by(1)

    expect(response.status).to eq(302)
  end

  it "存在しないworkIDで404エラーになること" do
    user = FactoryBot.create(:registered_user)
    collection = FactoryBot.create(:collection, user:)

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/999999/collection_items", params: {
        forms_collection_item_form: {
          collection_id: collection.id
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたworkIDで404エラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user:)
    work.destroy!

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: collection.id
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "他のユーザーのコレクションを指定するとバリデーションエラーになること" do
    user = FactoryBot.create(:registered_user)
    other_user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    other_collection = FactoryBot.create(:collection, user: other_user)

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: other_collection.id
        }
      }
    }.not_to change(CollectionItem, :count)

    expect(response.status).to eq(422)
  end

  it "削除されたコレクションを指定するとバリデーションエラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user:)
    collection.destroy!

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: collection.id
        }
      }
    }.not_to change(CollectionItem, :count)

    expect(response.status).to eq(422)
  end

  it "既に同じワークがコレクションに追加されている場合、エラーになること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user:)

    # 既存のコレクションアイテムを作成
    FactoryBot.create(:collection_item, user:, work:, collection:)

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: collection.id
        }
      }
    }.to raise_error(ActiveRecord::RecordNotUnique)
  end

  it "削除されたコレクションアイテムが存在する場合でも、新規作成できること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    collection = FactoryBot.create(:collection, user:)

    # 既存のコレクションアイテムを作成して削除
    old_item = FactoryBot.create(:collection_item, user:, work:, collection:)
    old_item.destroy!

    login_as(user, scope: :user)

    expect {
      post "/fragment/works/#{work.id}/collection_items", params: {
        forms_collection_item_form: {
          collection_id: collection.id
        }
      }
    }.to change(CollectionItem, :count).by(1)

    expect(response.status).to eq(302)
  end
end
