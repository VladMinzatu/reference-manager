package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/VladMinzatu/reference-manager/adapters"
	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/service"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "refman",
	Short: "Reference Manager CLI",
	Long:  "Command line interface for managing references organized by categories",
}

func main() {
	db, err := sql.Open("sqlite3", "db/references.db")
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()

	repo := adapters.NewSQLiteRepository(db)
	svc := service.NewReferenceService(repo)

	// Category commands
	var categoryCmd = &cobra.Command{
		Use:   "category",
		Short: "Manage categories",
	}

	var addCategoryCmd = &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cat, err := svc.AddCategory(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Added category: %s (id: %d)\n", cat.Name, cat.Id)
			return nil
		},
	}

	var listCategoriesCmd = &cobra.Command{
		Use:   "list",
		Short: "List all categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			categories, err := svc.GetAllCategories()
			if err != nil {
				return err
			}
			for _, cat := range categories {
				fmt.Printf("%d: %s\n", cat.Id, cat.Name)
			}
			return nil
		},
	}
	var updateCategoryCmd = &cobra.Command{
		Use:   "update [id] [new_name]",
		Short: "Update the name of a category",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id format (must be integer): %v", err)
			}
			newName := args[1]
			if err := svc.UpdateCategory(id, newName); err != nil {
				return err
			}
			fmt.Printf("Updated category %d to name: %s\n", id, newName)
			return nil
		},
	}

	var deleteCategoryCmd = &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id format (must be integer): %v", err)
			}
			if err := svc.DeleteCategory(id); err != nil {
				return err
			}
			fmt.Printf("Deleted category with id: %d\n", id)
			return nil
		},
	}

	var reorderCategoriesCmd = &cobra.Command{
		Use:   "reorder [id1] [id2] ...",
		Short: "Reorder categories by specifying their ids in the desired order",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			positions := make(map[int64]int)
			for pos, arg := range args {
				id, err := strconv.ParseInt(arg, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid category id: %s", arg)
				}
				positions[id] = pos
			}
			if err := svc.ReorderCategories(positions); err != nil {
				return err
			}
			fmt.Println("Categories reordered successfully.")
			return nil
		},
	}

	// Reference commands
	var referenceCmd = &cobra.Command{
		Use:   "reference",
		Short: "Manage references",
	}

	var listReferencesCmd = &cobra.Command{
		Use:   "list [categoryId] [starredOnly]",
		Short: "List references in a category, optionally filtering by starred references",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			starredOnly := false
			if len(args) == 2 {
				starredOnly, err = strconv.ParseBool(args[1])
				if err != nil {
					return fmt.Errorf("invalid starredOnly value (must be true or false): %v", err)
				}
			}
			refs, err := svc.GetReferences(categoryId, starredOnly)
			if err != nil {
				return err
			}
			for _, ref := range refs {
				ref.Render(&CLIReferenceRenderer{})
				fmt.Println()
			}
			return nil
		},
	}

	var addBookCmd = &cobra.Command{
		Use:   "add-book [categoryId] [title] [isbn] [description]",
		Short: "Add a book reference",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			book, err := svc.AddBookReference(categoryId, args[1], args[2], args[3])
			if err != nil {
				return err
			}
			fmt.Printf("Added book: %s (id: %d)\n", book.Title, book.Id)
			return nil
		},
	}

	var updateBookCmd = &cobra.Command{
		Use:   "update-book [id] [title] [isbn] [description] [starred]",
		Short: "Update a book reference",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid book id: %v", err)
			}
			title := args[1]
			isbn := args[2]
			description := args[3]
			starred, err := strconv.ParseBool(args[4])
			if err != nil {
				return fmt.Errorf("invalid starred value (must be true or false): %v", err)
			}
			if err := svc.UpdateBookReference(id, title, isbn, description, starred); err != nil {
				return err
			}
			fmt.Printf("Updated book (id: %d)\n", id)
			return nil
		},
	}

	var addLinkCmd = &cobra.Command{
		Use:   "add-link [categoryId] [title] [url] [description]",
		Short: "Add a link reference",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			link, err := svc.AddLinkReference(categoryId, args[1], args[2], args[3])
			if err != nil {
				return err
			}
			fmt.Printf("Added link: %s (id: %d)\n", link.Title, link.Id)
			return nil
		},
	}

	var updateLinkCmd = &cobra.Command{
		Use:   "update-link [id] [title] [url] [description] [starred]",
		Short: "Update a link reference",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid link id: %v", err)
			}
			title := args[1]
			url := args[2]
			description := args[3]
			starred, err := strconv.ParseBool(args[4])
			if err != nil {
				return fmt.Errorf("invalid starred value (must be true or false): %v", err)
			}
			if err := svc.UpdateLinkReference(id, title, url, description, starred); err != nil {
				return err
			}
			fmt.Printf("Updated link (id: %d)\n", id)
			return nil
		},
	}

	var addNoteCmd = &cobra.Command{
		Use:   "add-note [categoryId] [title] [text]",
		Short: "Add a note reference",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			note, err := svc.AddNoteReference(categoryId, args[1], args[2])
			if err != nil {
				return err
			}
			fmt.Printf("Added note: %s (id: %d)\n", note.Title, note.Id)
			return nil
		},
	}

	var updateNoteCmd = &cobra.Command{
		Use:   "update-note [id] [title] [text] [starred]",
		Short: "Update a note reference",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid note id: %v", err)
			}
			title := args[1]
			text := args[2]
			starred, err := strconv.ParseBool(args[3])
			if err != nil {
				return fmt.Errorf("invalid starred value (must be true or false): %v", err)
			}
			if err := svc.UpdateNoteReference(id, title, text, starred); err != nil {
				return err
			}
			fmt.Printf("Updated note (id: %d)\n", id)
			return nil
		},
	}

	var deleteReferenceCmd = &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a reference",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid reference id: %v", err)
			}
			if err := svc.DeleteReference(id); err != nil {
				return err
			}
			fmt.Printf("Deleted reference with id: %d\n", id)
			return nil
		},
	}

	var reorderReferencesCmd = &cobra.Command{
		Use:   "reorder [categoryId] [id1] [id2] ...",
		Short: "Reorder references in a category by specifying the ids in the desired order.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			positions := make(map[int64]int)
			for pos, idStr := range args[1:] {
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid reference id at position %d: %v", pos, err)
				}
				positions[id] = pos
			}
			if err := svc.ReorderReferences(categoryId, positions); err != nil {
				return err
			}
			fmt.Printf("Reordered references in category %d\n", categoryId)
			return nil
		},
	}

	categoryCmd.AddCommand(addCategoryCmd, listCategoriesCmd, updateCategoryCmd, deleteCategoryCmd, reorderCategoriesCmd)
	referenceCmd.AddCommand(listReferencesCmd, addBookCmd, updateBookCmd, addLinkCmd, updateLinkCmd, addNoteCmd, updateNoteCmd, deleteReferenceCmd, reorderReferencesCmd)
	rootCmd.AddCommand(categoryCmd, referenceCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type CLIReferenceRenderer struct{}

func (r *CLIReferenceRenderer) RenderBook(ref model.BookReference) {
	fmt.Printf("%d: %s [Book] %s\n", ref.Id, r.StarChar(ref.Starred), ref.Title)
	fmt.Printf("\t\t\tISBN: %s\n", ref.ISBN)
	fmt.Printf("\t\t\tDescription: %s\n", ref.Description)
}

func (r *CLIReferenceRenderer) RenderLink(ref model.LinkReference) {
	fmt.Printf("%d: %s [Link] %s\n", ref.Id, r.StarChar(ref.Starred), ref.Title)
	fmt.Printf("\t\t\tURL: %s\n", ref.URL)
	fmt.Printf("\t\t\tDescription: %s\n", ref.Description)
}

func (r *CLIReferenceRenderer) RenderNote(ref model.NoteReference) {
	fmt.Printf("%d: %s [Note] %s\n", ref.Id, r.StarChar(ref.Starred), ref.Title)
	fmt.Printf("\t\t\tText: %s\n", ref.Text)
}

func (r *CLIReferenceRenderer) StarChar(starred bool) string {
	if starred {
		return "★"
	}
	return "☆"
}
